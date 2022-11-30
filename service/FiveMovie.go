package service

import "C"
import (
	"bytes"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"log"
	"strconv"
	"strings"
	"time"
)

var (
	fiveHost      = "https://555movie.me"
	fiveTagUrl    = "https://555movie.me/label/netflix/page/%d.html"
	fiveSearchUrl = "https://555movie.me/vodsearch/%s----------%d---.html"
	fiveDetailUrl = "https://555movie.me/voddetail/%s.html"
	fivePlayUrl   = "https://555movie.me/vodplay/%s.html"
)

//========================================================================
//==============================接口实现===================================
//========================================================================

type FiveMovie struct{ Movie }

func (x FiveMovie) ListByTag(tagName, page string) model.Pager {
	return x.fiveListByTag(tagName, page)
}

func (x FiveMovie) Search(search, page string) model.Pager {
	return x.fiveListBySearch(search, page)
}

func (x FiveMovie) Detail(id string) model.MovieInfo {
	return x.fiveVideoDetail(id)
}

func (x FiveMovie) Source(sid, vid string) model.Video {
	return x.fiveVideoSource(sid, vid)
}

//========================================================================
//==============================实际业务处理逻辑============================
//========================================================================

func (x FiveMovie) fiveListByTag(tagName, page string) model.Pager {
	_page, _ := strconv.Atoi(page)

	var pager = model.Pager{}
	pager.Limit = 16

	c := x.Movie.NewColly()

	c.OnHTML(".tab-list .module-items a", func(element *colly.HTMLElement) {
		name := element.ChildText(".module-poster-item-title")
		url := element.Attr("href")
		thumb := element.ChildAttr("img.lazy", "data-original")
		tag := element.ChildText(".module-item-note")
		actors := element.ChildText(".xxx")
		resolution := element.ChildText(".module-item-note")

		pager.List = append(pager.List, model.MovieInfo{
			Id:         util.CZHandleUrlToId(url),
			Name:       name,
			Thumb:      thumb,
			Url:        url,
			Actors:     actors,
			Tag:        tag,
			Resolution: resolution,
		})
	})

	c.OnHTML("#page", func(element *colly.HTMLElement) {
		element.ForEach("a.page-next", func(i int, element *colly.HTMLElement) {
			tmpList := strings.Split(element.Attr("href"), "/")
			n, _ := strconv.Atoi(strings.TrimRight(tmpList[len(tmpList)-1], ".html"))
			if n*pager.Limit > pager.Total {
				pager.Total = n * pager.Limit
			}
		})
	})

	c.OnHTML("#page .page-current", func(element *colly.HTMLElement) {
		pager.Current, _ = strconv.Atoi(element.Text)
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})
	c.OnResponse(func(response *colly.Response) {
		if newResp := isWaf(string(response.Body)); newResp != nil {
			response.Body = newResp
		}
	})

	err := c.Visit(fmt.Sprintf(fiveTagUrl, _page))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return pager
}

func (x FiveMovie) fiveListBySearch(query, page string) model.Pager {
	var requestUrl = fmt.Sprintf(fiveSearchUrl, query, util.HandlePageNumber(page))
	var pager = model.Pager{}
	pager.Limit = 16

	c := x.Movie.NewColly()

	c.OnHTML(".module-items .module-card-item", func(element *colly.HTMLElement) {
		name := element.ChildText(".module-card-item-title a strong")
		url := element.ChildAttr(".module-card-item-title a", "href")
		thumb := element.ChildAttr(".module-item-pic .lazyload", "data-original")
		tag := element.ChildText(".module-item-note")
		actors := element.ChildText(".module-info-item-content")

		pager.List = append(pager.List, model.MovieInfo{
			Id:     util.CZHandleUrlToId(url),
			Name:   name,
			Thumb:  thumb,
			Url:    url,
			Actors: actors,
			Tag:    tag,
		})
	})

	c.OnHTML("#page", func(element *colly.HTMLElement) {
		// /vodsearch/WE----------51---.html
		element.ForEach("a", func(i int, element *colly.HTMLElement) {

			tmpList := strings.Split(element.Attr("href"), "-")
			if len(tmpList) < 4 {
				return
			}
			pager.Total, _ = strconv.Atoi(tmpList[len(tmpList)-4])
		})
	})

	pager.Current = util.HandlePageNumber(page)

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})
	c.OnResponse(func(response *colly.Response) {
		log.Println("[responseBody]", string(response.Body))
		x.checkFiveWaf(requestUrl, response.Body)
	})

	err := c.Visit(requestUrl)
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return pager
}

func (x FiveMovie) fiveVideoDetail(id string) model.MovieInfo {
	var info = model.MovieInfo{Id: id}
	var idxList = make([][]string, 0)

	c := x.Movie.NewColly()

	c.OnHTML(".module-tab-items-box", func(element *colly.HTMLElement) {
		element.ForEach(".module-tab-item", func(i0 int, element *colly.HTMLElement) {
			idxList = append(idxList, []string{strconv.Itoa(i0), element.ChildText("span")})
		})
	})

	c.OnHTML(".module-info-heading", func(element *colly.HTMLElement) {
		info.Name = element.ChildText("h1")
	})
	c.OnHTML(".module-info-poster .module-item-pic .lazyload", func(element *colly.HTMLElement) {
		info.Thumb = element.Attr("data-original")
	})
	c.OnHTML(".module-info-content .module-info-introduction-content p", func(element *colly.HTMLElement) {
		info.Intro = strings.TrimSpace(element.Text)
	})

	c.OnHTML("body", func(element *colly.HTMLElement) {
		element.ForEach(".module-list", func(i int, element *colly.HTMLElement) {
			for _, tmpValue := range idxList {
				if strconv.Itoa(i) != tmpValue[0] {
					continue
				}
				element.ForEach(".module-play-list-link", func(i int, element *colly.HTMLElement) {
					info.Links = append(info.Links, model.Link{
						Id:    util.CZHandleUrlToId2(element.Attr("href")),
						Name:  element.ChildText("span"),
						Url:   element.Attr("href"),
						Group: tmpValue[1],
					})
				})
			}
		})
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf(fiveDetailUrl, id))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return info
}

func (x FiveMovie) fiveVideoSource(sid, vid string) model.Video {
	var video = model.Video{Id: sid}

	video.Source = x.fiveParseVideoUrl(fmt.Sprintf(fivePlayUrl, sid))

	video.Url = HandleSrcM3U8FileToLocal(sid, video.Source, x.Movie.IsCache)

	// 视频类型问题处理
	video = x.handleVideoType(video)

	return video
}

func (x FiveMovie) handleVideoType(v model.Video) model.Video {
	v.Type = "hls"
	return v
}

// 筛选网络请求，找到特定地址（可能是播放地址）返回
func (x FiveMovie) fiveParseVideoUrl(requestUrl string) string {
	var findUrl string

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			//log.Println("[network.EventRequestWillBeSent]", ev.Type, ev.Request.URL)
			if util.StringInList(ev.Type.String(), []string{"Stylesheet", "Image", "Font"}) {
				ev.Request.URL = ""
			}
			if util.StringInList(util.HandleHost(ev.Request.URL), util.FiveVideoHost) {
				findUrl = ev.Request.URL
				cancel()
			}
		case *network.EventWebSocketCreated:
			//log.Println("[network.EventWebSocketCreated]", ev.URL)
		case *network.EventWebSocketFrameError:
			//log.Println("[network.EventWebSocketFrameError]", ev.ErrorMessage)
		case *network.EventWebSocketFrameSent:
			//log.Println("[network.EventWebSocketFrameSent]", ev.Response.PayloadData)
		case *network.EventWebSocketFrameReceived:
			//log.Println("[network.EventWebSocketFrameReceived]", ev.Response.PayloadData)
		case *network.EventResponseReceived:
			//log.Println("[network.EventResponseReceived]", ev.Type, ev.Response.URL)
		}
	})

	err := chromedp.Run(ctx,
		chromedp.Tasks{
			network.Enable(),
			chromedp.Navigate(requestUrl),
			chromedp.WaitVisible("#I_FUCK_YOU"), // 等一个不存在的节点，然后通过event中cancel()接下来的所有request
		},
	)

	if err != nil {
		log.Println("[chromedp.Run.Error]", err.Error())
	}

	return findUrl
}

// 认证页面模拟点击一下
func (x FiveMovie) fiveGuardClick(requestUrl string) string {
	var findUrl string

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			//log.Println("[network.EventRequestWillBeSent]", ev.Type, ev.Request.URL)
			if util.StringInList(ev.Type.String(), []string{"Stylesheet", "Image", "Font"}) {
				ev.Request.URL = ""
			}
		case *network.EventWebSocketCreated:
			//log.Println("[network.EventWebSocketCreated]", ev.URL)
		case *network.EventWebSocketFrameError:
			//log.Println("[network.EventWebSocketFrameError]", ev.ErrorMessage)
		case *network.EventWebSocketFrameSent:
			//log.Println("[network.EventWebSocketFrameSent]", ev.Response.PayloadData)
		case *network.EventWebSocketFrameReceived:
			//log.Println("[network.EventWebSocketFrameReceived]", ev.Response.PayloadData)
		case *network.EventResponseReceived:
			//log.Println("[network.EventResponseReceived]", ev.Type, ev.Response.URL)
		}
	})

	//var res []byte
	err := chromedp.Run(ctx,
		chromedp.Tasks{
			network.Enable(),
			chromedp.Navigate(requestUrl),
			chromedp.WaitVisible("#access"),
			chromedp.Click("#access"),
			chromedp.WaitVisible(".module-heading-search"),
			//chromedp.FullScreenshot(&res, 90),
		},
	)

	if err != nil {
		log.Println("[chromedp.Run.Error]", err.Error())
	}
	//if err := os.WriteFile("fullScreenshot.png", res, fs.ModePerm); err != nil {
	//	log.Fatal(err)
	//}

	return findUrl
}

func (x FiveMovie) checkFiveWaf(requestUrl string, responseBody []byte) []byte {
	// <script src="/_guard/html.js?js=click_html"></script>
	if !bytes.Contains(responseBody, []byte("_guard/html.js?js=click_html")) {
		return responseBody
	}

	x.fiveGuardClick(requestUrl)

	return nil
}
