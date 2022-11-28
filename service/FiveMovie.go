package service

import "C"
import (
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	fiveHost      = "https://555movie.me"
	fiveTagUrl    = "https://555movie.me/label/new/page/%d.html"
	fiveSearchUrl = "https://www.czspp.com/xssearch?q=%s&p=%d"
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

	c.OnHTML(".module-main #page", func(element *colly.HTMLElement) {
		element.ForEach("a.page-next", func(i int, element *colly.HTMLElement) {
			tmpList := strings.Split(element.Attr("href"), "/")
			n, _ := strconv.Atoi(tmpList[len(tmpList)-1])
			if n*pager.Limit > pager.Total {
				pager.Total = n * pager.Limit
			}
		})
	})

	c.OnHTML(".module-main .page-current", func(element *colly.HTMLElement) {
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

	log.Println(fmt.Sprintf(czTagUrl, tagName, _page))

	err := c.Visit(fmt.Sprintf(fiveTagUrl, _page))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return pager
}

func (x FiveMovie) fiveListBySearch(query, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 20

	c := x.Movie.NewColly()

	c.OnHTML(".search_list ul li", func(element *colly.HTMLElement) {
		name := element.ChildText(".dytit a")
		url := element.ChildAttr(".dytit a", "href")
		thumb := element.ChildAttr("img.thumb", "data-original")
		tag := element.ChildText(".nostag")
		actors := element.ChildText(".inzhuy")

		pager.List = append(pager.List, model.MovieInfo{
			Id:     util.CZHandleUrlToId(url),
			Name:   name,
			Thumb:  thumb,
			Url:    url,
			Actors: actors,
			Tag:    tag,
		})
	})

	c.OnHTML(".dytop .dy_tit_big", func(element *colly.HTMLElement) {
		element.ForEach("span", func(i int, element *colly.HTMLElement) {
			if i == 0 {
				pager.Total, _ = strconv.Atoi(element.Text)
			}
		})
	})

	c.OnHTML(".pagenavi_txt .current", func(element *colly.HTMLElement) {
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

	err := c.Visit(fmt.Sprintf(czSearchUrl, query, util.HandlePageNumber(page)))
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

	video.Source = x.fiveParseVideoUrl(sid)
	video.Url = video.Source

	// 视频类型问题处理
	video = x.handleVideoType(video)

	return video
}

func (x FiveMovie) handleVideoType(v model.Video) model.Video {
	v.Type = "hls"
	return v
}

func (x FiveMovie) fiveParseVideoSource(id, js string) (model.Video, error) {
	var video = model.Video{}
	tmpList := strings.Split(strings.TrimSpace(js), ";")

	var data = ""
	var key = ""
	var iv = ""
	for index, str := range tmpList {
		if index == 0 {
			regex := regexp.MustCompile(`"\S+"`)
			data = strings.Trim(regex.FindString(str), `"`)
			continue
		}
		if index == 1 {
			regex := regexp.MustCompile(`"(\S+)"`)
			matchList := regex.FindStringSubmatch(str)
			if len(matchList) > 0 {
				key = matchList[len(matchList)-1]
			}
			continue
		}
		if index == 2 {
			regex := regexp.MustCompile(`\((\S+)\)`)
			matchList := regex.FindStringSubmatch(str)
			if len(matchList) > 0 {
				iv = matchList[len(matchList)-1]
			}
			continue
		}
	}

	log.Println(fmt.Sprintf("[parsing] key: %s, iv: %s", key, iv))

	if key == "" && data == "" {
		return video, errors.New("解析失败")
	}
	bs, err := util.DecryptByAes([]byte(key), []byte(iv), data)
	if err != nil {
		return video, errors.New("解密失败")
	}
	tmpList = strings.Split(string(bs), "window")
	if len(tmpList) < 1 {
		return video, errors.New("解密数据错误")
	}

	regex := regexp.MustCompile(`{url: "(\S+)",type:"(\S+)",([\S\s]*)pic:'(\S+)'}`)
	matchList := regex.FindStringSubmatch(tmpList[0])

	if len(matchList) < 1 {
		return video, errors.New("解析视频信息失败")
	}

	video.Id = id

	for index, m := range matchList {
		switch index {
		case 1:
			video.Source = m
			video.Url = m
			break
		case 2:
			video.Type = m
			break
		case 4:
			video.Thumb = m
			break
		default:
			break
		}
	}

	video.Url = HandleSrcM3U8FileToLocal(id, video.Source, x.Movie.IsCache)

	return video, nil
}

func (x FiveMovie) fiveParseVideoUrl(id string) string {
	var findUrl string

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			log.Println("[network.EventRequestWillBeSent]", ev.Type, ev.Request.URL)
			if util.StringInList(ev.Type.String(), []string{"Stylesheet", "Image", "Font"}) {
				ev.Request.URL = ""
			}
			log.Println("[network.EventRequestWillBeSent]", ev.Type, util.HandleHost(ev.Request.URL))
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
			chromedp.Navigate(fmt.Sprintf(fivePlayUrl, id)),
			chromedp.WaitVisible("#I_FUCK_YOU"), // 等一个不存在的节点，然后通过event中cancel()接下来的所有request
		},
	)

	if err != nil {
		log.Println("[chromedp.Run.Error]", err.Error())
	}

	return findUrl
}
