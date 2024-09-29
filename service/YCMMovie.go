package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"github.com/zc310/headers"
	"io/ioutil"
	"log"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	// https://www.ycmdy.com/
	ycmHost      = "https://www.kanjugg.com"
	ycmTagUrl    = "https://www.kanjugg.com/list/?1-%d.html"
	ycmSearchUrl = "https://www.kanjugg.com/search.php?searchword=我的"
	ycmDetailUrl = "https://www.kanjugg.com/detail/?%s.html"
	ycmPlayUrl   = "https://www.kanjugg.com/video/?%s.html"
)

//========================================================================
//==============================接口实现===================================
//========================================================================

type YCMMovie struct {
	movie       Movie
	httpWrapper *util.HttpWrapper
	btVerifyUrl string
}

func (x *YCMMovie) Init(movie Movie) {
	x.movie = movie
	if x.httpWrapper == nil {
		x.httpWrapper = &util.HttpWrapper{}
	}
	x.httpWrapper.SetHeader(headers.Origin, ycmHost)
	x.httpWrapper.SetHeader(headers.Referer, ycmHost)
	x.httpWrapper.SetHeader(headers.UserAgent, ua)
}

func (x *YCMMovie) ListByTag(tagName, page string) model.Pager {
	return x.ysListByTag(tagName, page)
}

func (x *YCMMovie) Search(search, page string) model.Pager {
	return x.ysListBySearch(search, page)
}

func (x *YCMMovie) Detail(id string) model.MovieInfo {
	return x.ysVideoDetail(id)
}

func (x *YCMMovie) Source(sid, vid string) model.Video {
	return x.ysVideoSource(sid, vid)
}

//========================================================================
//==============================实际业务处理逻辑============================
//========================================================================

func (x *YCMMovie) ysListByTag(tagName, page string) model.Pager {
	_page, _ := strconv.Atoi(page)

	var pager = model.Pager{}
	pager.Limit = 10

	b, err := x.httpWrapper.Get(fmt.Sprintf(ycmTagUrl, _page))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return pager
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return pager
	}

	var totalPageString = strings.TrimSpace(doc.Find(".myui-page .visible-xs ").Text())
	pager.Total = util.StringToInt(util.SimpleRegEx(totalPageString, `\/(\d+)`))*pager.Limit + 1

	doc.Find(".myui-panel .flickity .myui-vodlist__box").Each(func(i int, selection *goquery.Selection) {
		name := selection.Find(".title").Text()
		tmpUrl, _ := selection.Find(".myui-vodlist__thumb").Attr("href")
		tmpStyle, _ := selection.Find(".myui-vodlist__thumb").Attr("style")
		tag := selection.Find(".pic-tag").Text()
		actors := selection.Find(".text-muted").Text()
		resolution := selection.Find(".pic-text").Text()

		pager.List = append(pager.List, model.MovieInfo{
			Id:         util.SimpleRegEx(tmpUrl, `(\d+)`),
			Name:       name,
			Thumb:      util.SimpleRegEx(tmpStyle, `background: url\((\S+)\);`),
			Url:        tmpUrl,
			Actors:     strings.TrimSpace(actors),
			Tag:        strings.TrimSpace(tag),
			Resolution: strings.TrimSpace(resolution),
		})
	})

	log.Println("[pager.List]", len(pager.List))

	return pager
}

func (x *YCMMovie) ysListBySearch(query, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 20

	b, err := x.httpWrapper.Get(fmt.Sprintf(ycmSearchUrl))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return pager
	}
	var respHtml = string(b)

	if strings.Contains(respHtml, "window._cf_chl_opt") {
		log.Println("【cloudflare】waf")
		x.getHtmlCrossCloudflare(ycmSearchUrl)

	}

	return pager
}

func (x *YCMMovie) ysVideoDetail(id string) model.MovieInfo {
	var info = model.MovieInfo{Id: id}

	b, err := x.httpWrapper.Get(fmt.Sprintf(ycmDetailUrl, id))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return info
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return info
	}

	info.Name, _ = doc.Find(".myui-content__thumb .myui-vodlist__thumb").Attr("title")
	info.Thumb, _ = doc.Find(".myui-content__thumb .lazyload").Attr("data-original")
	info.Thumb = util.FillUrlHost(info.Thumb, ycmHost)
	info.Intro = strings.TrimSpace(doc.Find(".myui-content__detail .desc .data").Text())
	info.Intro = strings.ReplaceAll(info.Intro, " ", "")

	var groupList = make([]string, 0)
	doc.Find(".myui-panel-box .nav-tabs li").Each(func(i int, selection *goquery.Selection) {
		groupList = append(groupList, selection.Text())
	})

	log.Println("[groupList]", util.ToJSON(groupList, true))

	doc.Find(".myui-panel-box .tab-content .sort-list").Each(func(i int, selection *goquery.Selection) {
		var tmpGroup = groupList[i]
		selection.Find(".myui-content__list li").Each(func(i int, selection *goquery.Selection) {
			tmpUrl, _ := selection.Find("a").Attr("href")
			tmpName, _ := selection.Find("a").Attr("title")
			info.Links = append(info.Links, model.Link{
				Id:    util.SimpleRegEx(tmpUrl, `(\d+-\d+-\d+)`),
				Name:  tmpName,
				Url:   tmpUrl,
				Group: tmpGroup,
			})
		})
	})

	return info
}

func (x *YCMMovie) ysVideoSource(sid, vid string) model.Video {
	var video = model.Video{Id: sid, Vid: vid}

	b, err := x.httpWrapper.Get(fmt.Sprintf(ycmPlayUrl, sid))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return video
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return video
	}

	video.Name = doc.Find(".myui-panel .col-pd .myui-panel__head .title").Text()
	video.Thumb, _ = doc.Find(".vod_history").Attr("data-pic")
	video.Thumb = util.FillUrlHost(video.Thumb, ycmHost)

	var playJsText = doc.Find(".embed-responsive").Text()

	var findEncodedUrl = util.SimpleRegEx(playJsText, `now=base64decode\(\"(\S+)\"\);`)
	tmpBuff, _ := base64.StdEncoding.DecodeString(findEncodedUrl)
	video.Source = findEncodedUrl
	video.Url = string(tmpBuff)
	video.Type = util.GuessVideoType(video.Url)

	return video
}

func (x *YCMMovie) getHtmlCrossCloudflare(requestUrl string) string {
	var findUrl string

	allocCtx, allocCancel := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.UserAgent(ua),
	)
	defer allocCancel()

	ctx, ctxCancel := chromedp.NewContext(allocCtx)
	defer ctxCancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, timeoutCancel := context.WithTimeout(ctx, 80*time.Second)
	defer timeoutCancel()

	//var respHtml string
	var urlMap = map[network.RequestID]string{}

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			//log.Println("[network.EventRequestWillBeSent]", ev.Type, ev.Request.URL)
		case *network.EventWebSocketCreated:
			//log.Println("[network.EventWebSocketCreated]", ev.URL)
		case *network.EventWebSocketFrameError:
			log.Println("[network.EventWebSocketFrameError]", ev.ErrorMessage)
		case *network.EventWebSocketFrameSent:
			//log.Println("[network.EventWebSocketFrameSent]", ev.Response.PayloadData)
		case *network.EventWebSocketFrameReceived:
			//log.Println("[network.EventWebSocketFrameReceived]", ev.Response.PayloadData)
		case *network.EventResponseReceived:
			log.Println("[network.EventResponseReceived]", ev.Type, ev.RequestID, ev.Response.URL)
			if ev.Type == network.ResourceTypeDocument {
				urlMap[ev.RequestID] = ev.Response.URL
			}
		case *network.EventLoadingFinished:
			//urlMap[ev.RequestID] = ev.Response.URL
			if _, ok := urlMap[ev.RequestID]; ok {
				go func() {
					// print response body
					c := chromedp.FromContext(ctx)
					rbp := network.GetResponseBody(ev.RequestID)
					body, err := rbp.Do(cdp.WithExecutor(ctx, c.Target))
					if err != nil {
						log.Println("[network.body.WithExecutor.ErrorF]", ev.RequestID, err.Error())
					}
					if err = ioutil.WriteFile(ev.RequestID.String(), body, 0644); err != nil {
						log.Println("[network.body.WriteFile.ErrorF]", ev.RequestID, err.Error())
					}
					if err == nil {
						//log.Println("[network.body.F]", ev.RequestID, string(body))
						log.Println("[network.body.Full]", ev.RequestID, len(string(body)))
						//respHtml += string(body)
					}

				}()
			}
		case *runtime.EventConsoleAPICalled:
		case *fetch.EventRequestPaused:
			log.Println("[fetch.EventRequestPaused]", ev.RequestID, ev.ResourceType, ev.Request.Method, ev.Request.URL)

			go func() {
				parsed, err := url.Parse(ev.Request.URL)
				if err != nil {
					return
				}
				var isChallenge = strings.HasPrefix(parsed.Path, "/cdn-cgi/challenge-platform")
				var isTurnstile = strings.HasPrefix(parsed.Path, "/turnstile/")

				c := chromedp.FromContext(ctx)
				ctx := cdp.WithExecutor(ctx, c.Target)
				if slices.Contains([]network.ResourceType{network.ResourceTypeStylesheet, network.ResourceTypeImage, network.ResourceTypeFont}, ev.ResourceType) {
					_ = fetch.FailRequest(ev.RequestID, network.ErrorReasonConnectionAborted).Do(ctx)
				} else if ev.ResourceType == network.ResourceTypeScript && (!isChallenge && !isTurnstile) {
					_ = fetch.FailRequest(ev.RequestID, network.ErrorReasonConnectionAborted).Do(ctx)
				} else {
					_ = fetch.ContinueRequest(ev.RequestID).Do(ctx)
				}
			}()

		}
	})

	var respHtml = ""
	//var isWebDriver bool
	//var screenshot []byte
	//var cookie string
	err := chromedp.Run(
		ctx,
		fetch.Enable(),
		network.Enable(),
		chromedp.EmulateViewport(880, 435),
		chromedp.Navigate(requestUrl),
		chromedp.Tasks{
			chromedp.WaitVisible(".myui-vodlist__media"),
			chromedp.WaitVisible(".myui-page"),
			chromedp.InnerHTML("html", &respHtml),
			chromedp.Sleep(time.Second * 2),
		},
	)

	if err != nil {
		log.Println("[chromedp.Run.Error]", err.Error())
	}

	log.Println("[respHtml===========>]", respHtml)

	return findUrl
}

func (x *YCMMovie) parseCookie(cookies []*network.Cookie) string {
	log.Println("[parseCookie]", util.ToJSON(cookies, true))
	var cookieString = ""
	if cookies == nil || len(cookies) <= 0 {
		return cookieString
	}
	for _, cookie := range cookies {
		if len(cookieString) <= 0 {
			cookieString = fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
		} else {
			cookieString = fmt.Sprintf("%s; %s=%s", cookieString, cookie.Name, cookie.Value)
		}
	}
	return cookieString
}
