package service

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"log"
	"regexp"
	"strings"
	"time"
)

var (
	lvSearchUrl = "https://lvv2.com/mv/search/q-%s/type-/page-%d"
	lvDetailUrl = "https://lvv2.com/mv/play/id-%s"
)

type LVMovie struct{ Movie }

func (x LVMovie) ListByTag(tagName, page string) model.Pager {
	tagName = "天"
	return x.lvListBySearch(tagName, page)
}

func (x LVMovie) Search(search, page string) model.Pager {
	return x.lvListBySearch(search, page)
}

func (x LVMovie) Detail(id string) model.MovieInfo {
	return x.lvVideoDetail(id)
}

func (x LVMovie) Source(sid, vid string) model.Video {
	return x.lvVideoSource2(sid, vid)
}

// ===

func (x LVMovie) lvHandleUrlToId(tmpUrl string) int {
	tmpList := strings.Split(tmpUrl, "/")
	return util.HandlePageNumber(strings.TrimLeft(tmpList[len(tmpList)-1], "page-"))
}

func (x LVMovie) lvListBySearch(query, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 24
	pager.Current = util.HandlePageNumber(page)

	c := x.Movie.NewColly()

	c.OnHTML(".wrap-content .stui-pannel-box", func(element *colly.HTMLElement) {
		name := element.ChildText(".stui-content__detail .title a")
		url := element.ChildAttr(".stui-content__detail .title a", "href")
		thumb := element.ChildAttr(".stui-content__thumb img", "data-original")
		tag := ""
		actors := ""

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
		element.ForEach("a", func(i int, element *colly.HTMLElement) {
			pager.Total = pager.Limit * x.lvHandleUrlToId(element.Attr("href"))
		})
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf(lvSearchUrl, query, pager.Current))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return pager
}

func (x LVMovie) lvVideoDetail(id string) model.MovieInfo {
	var info = model.MovieInfo{}
	info.Id = id

	c := x.Movie.NewColly()

	c.OnHTML(".stui-player__detail", func(element *colly.HTMLElement) {
		info.Name = strings.TrimSpace(element.ChildText(".title"))
		info.Intro = strings.TrimPrefix(element.ChildText(".desc"), "简介：")
	})
	c.OnResponse(func(response *colly.Response) {
		// urls[1][0]='https://v10.dious.cc/20210707/jxnrMrb8/index.m3u8';
		reg := regexp.MustCompile(`urls\[(\d+)\]\[(\d+)\]='(\S+)'`)
		urlList := reg.FindAllStringSubmatch(string(response.Body), -1)
		for _, urls := range urlList {
			if len(urls) != 4 {
				continue
			}
			info.Links = append(info.Links, model.Link{
				Id:    fmt.Sprintf("%s-%s", urls[1], urls[2]),
				Name:  fmt.Sprintf("资源%s", urls[2]),
				Url:   urls[3],
				Group: fmt.Sprintf("group_%s", urls[1]),
			})
		}
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf(lvDetailUrl, id))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return info
}

func (x LVMovie) lvVideoSource(sid, vid string) model.Video {
	var info = model.Video{}
	info.Id = vid
	info.Source = fmt.Sprintf(lvDetailUrl, vid)

	c := x.Movie.NewColly()

	c.OnHTML(".stui-player__detail", func(element *colly.HTMLElement) {
		info.Name = strings.TrimSpace(element.ChildText(".title"))
	})
	c.OnResponse(func(response *colly.Response) {
		tmpList := strings.Split(sid, "-")
		log.Println("[tmpList]", util.ToJSON(tmpList, true))
		if len(tmpList) != 2 {
			return
		}
		// urls[1][0]='https://v10.dious.cc/20210707/jxnrMrb8/index.m3u8';
		reg := regexp.MustCompile(`urls\[(\d+)\]\[(\d+)\]='(\S+)'`)
		urlList := reg.FindAllStringSubmatch(string(response.Body), -1)
		log.Println("[urlList]", util.ToJSON(urlList, true))
		for _, urls := range urlList {
			if len(urls) != 4 {
				continue
			}
			if urls[1] == tmpList[0] && urls[2] == tmpList[1] {
				//info.Url = urls[3]
				//info.Url = HandleSrcM3U8FileToLocal(sid, urls[3], x.Movie.IsCache)
				info.Url = urls[3]
				info.Type = "hls"
			}
		}
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf(lvDetailUrl, vid))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return info
}

func (x LVMovie) lvVideoSource2(sid, vid string) model.Video {
	var info = model.Video{}
	info.Id = sid
	info.Source = fmt.Sprintf(lvDetailUrl, vid)

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			if util.StringInList(ev.Type.String(), []string{"Stylesheet", "Image", "Font"}) {
				ev.Request.URL = ""
			}
			if strings.HasSuffix(ev.Request.URL, ".m3u8") {
				log.Println("[FIND]", ev.Request.URL)
				tmpInfo := x.getVideoUrl(sid, ev.Request.URL)
				info.Url = tmpInfo.Url
				info.Type = "hls"
				log.Println("[tmpInfo.Url]", tmpInfo.Url)
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
			log.Println("[network.EventResponseReceived]", ev.Type, ev.Response.URL)
		}
	})

	//var res []byte
	err := chromedp.Run(ctx,
		chromedp.Tasks{
			network.Enable(),
			chromedp.Navigate(info.Source),
			chromedp.WaitVisible("#fuck_access"),
			//chromedp.FullScreenshot(&res, 90),
		},
	)

	if err != nil {
		log.Println("[chromedp.Run.Error]", err.Error())
	}

	return info
}

func (x LVMovie) getVideoUrl(sid, requestUrl string) model.Video {
	var info = model.Video{}
	info.Id = sid
	if util.StringInList(util.HandleHost(requestUrl), util.RedirectConfig) {
		// 拿到重定向URL
		requestUrl = util.HandleRedirectUrl(requestUrl)
	}
	// 下载m3u8
	//info.Url = HandleSrcM3U8FileToLocal(sid, requestUrl, x.Movie.IsCache)
	info.Url = requestUrl
	info.Type = "hls"

	return info
}
