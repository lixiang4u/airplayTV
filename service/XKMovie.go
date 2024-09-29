package service

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"github.com/tidwall/gjson"
	"github.com/zc310/headers"
	"log"
	"net/url"
	"strings"
)

var (
	xkHost      = "https://www.huibangpaint.com/"
	xkTagUrl    = "https://www.ytshenxian.com/yt/dianying/page/%d.html"
	xkSearchUrl = "https://www.ytshenxian.com/search.html"
	xkDetailUrl = "https://www.ytshenxian.com/yplay/%s.html"
	xkPlayUrl   = "https://www.ytshenxian.com/yplay/%s.html"
)

type XKMovie struct {
	movie       Movie
	httpWrapper *util.HttpWrapper
}

func (x *XKMovie) Init(movie Movie) {
	x.movie = movie
	if x.httpWrapper == nil {
		x.httpWrapper = &util.HttpWrapper{}
	}

	x.httpWrapper.SetHeader(headers.Origin, xkHost)
	x.httpWrapper.SetHeader(headers.Referer, xkHost)
	x.httpWrapper.SetHeader(headers.UserAgent, ua)
	x.httpWrapper.SetHeader(headers.UserAgent, ua)
}

func (x *XKMovie) ListByTag(tagName, page string) model.Pager {
	return x.nnListByTag("", page)
}

func (x *XKMovie) Search(search, page string) model.Pager {
	return x.nnListBySearch(search, page)
}

func (x *XKMovie) Detail(id string) model.MovieInfo {
	return x.nnVideoDetail(id)
}

func (x *XKMovie) Source(sid, vid string) model.Video {
	return x.nnVideoSource(sid, vid)
}

//========================================================================
//==============================实际业务处理逻辑============================
//========================================================================

func (x *XKMovie) nnListBySearch(search, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 10 // 每页10条

	x.httpWrapper.SetHeader("content-type", "application/x-www-form-urlencoded")
	b, err := x.httpWrapper.Post(xkSearchUrl, fmt.Sprintf("wd=%s", search))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return pager
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return pager
	}

	var totalPageString = strings.TrimSpace(doc.Find(".stui-page .num").Text())
	pager.Total = util.StringToInt(util.SimpleRegEx(totalPageString, `(\d+)`))*pager.Limit + 1

	doc.Find(".stui-vodlist .stui-vodlist__item").Each(func(i int, selection *goquery.Selection) {
		name := selection.Find(".stui-vodlist__title").Text()
		tmpUrl, _ := selection.Find(".stui-vodlist__thumb").Attr("href")
		thumb, _ := selection.Find(".lazyload").Attr("data-original")
		tag := selection.Find(".pic-text").Text()
		resolution := selection.Find(".class").Text()

		pager.List = append(pager.List, model.MovieInfo{
			Id:         util.CZHandleUrlToId2(tmpUrl),
			Name:       name,
			Thumb:      thumb,
			Url:        tmpUrl,
			Actors:     "",
			Tag:        strings.TrimSpace(tag),
			Resolution: resolution,
		})
	})

	return pager
}

func (x *XKMovie) nnListByTag(tagName, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 42
	pager.Current = util.HandlePageNumber(page)

	b, err := x.httpWrapper.Get(fmt.Sprintf(xkTagUrl, pager.Current))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return pager
	}
	var respHtml = string(b)
	var cfUrl = fmt.Sprintf(cloudflareUrl, url.QueryEscape(fmt.Sprintf(xkTagUrl, pager.Current)), url.QueryEscape(".stui-vodlist"))
	respHtml = fuckCloudflare(respHtml, cfUrl)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(respHtml))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return pager
	}

	var totalPageString = strings.TrimSpace(doc.Find(".stui-page .num").Text())
	pager.Total = util.StringToInt(util.SimpleRegEx(totalPageString, `(\d+)`))*pager.Limit + 1

	doc.Find(".stui-vodlist .stui-vodlist__item").Each(func(i int, selection *goquery.Selection) {
		name := selection.Find(".stui-vodlist__title").Text()
		tmpUrl, _ := selection.Find(".stui-vodlist__thumb").Attr("href")
		thumb, _ := selection.Find(".lazyload").Attr("data-original")
		tag := selection.Find(".pic-text").Text()
		resolution := selection.Find(".class").Text()

		pager.List = append(pager.List, model.MovieInfo{
			Id:         util.CZHandleUrlToId2(tmpUrl),
			Name:       name,
			Thumb:      thumb,
			Url:        tmpUrl,
			Actors:     "",
			Tag:        strings.TrimSpace(tag),
			Resolution: resolution,
		})
	})

	return pager
}

// 根据id获取视频播放列表信息
func (x *XKMovie) nnVideoDetail(id string) model.MovieInfo {
	var info = model.MovieInfo{Id: id}

	b, err := x.httpWrapper.Get(fmt.Sprintf(xkDetailUrl, id))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return info
	}

	var respHtml = string(b)
	var cfUrl = fmt.Sprintf(cloudflareUrl, url.QueryEscape(fmt.Sprintf(xkDetailUrl, id)), url.QueryEscape(".playlist .nav-tabs li"))
	respHtml = fuckCloudflare(respHtml, cfUrl)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(respHtml))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return info
	}

	info.Name, _ = doc.Find(".stui-content__desc").Find(".pic").Attr("title")
	info.Intro = strings.TrimSpace(doc.Find(".stui-content__desc").Find(".info").Text())
	info.Intro = strings.ReplaceAll(info.Intro, " ", "")
	info.Thumb, _ = doc.Find(".stui-content__desc").Find(".img-responsive").Attr("src")

	var groupList = make([]string, 0)
	doc.Find(".playlist .nav-tabs li").Each(func(i int, selection *goquery.Selection) {
		groupList = append(groupList, selection.Text())
	})

	doc.Find(".playlist .rb-item").Each(func(i int, selection *goquery.Selection) {
		var tmpGroup = groupList[i]
		selection.Find("ul li").Each(func(i int, selection *goquery.Selection) {
			tmpUrl, _ := selection.Find("a").Attr("href")
			info.Links = append(info.Links, model.Link{
				Id:    util.SimpleRegEx(tmpUrl, `(\d+-\d+-\d+)`),
				Name:  selection.Find("a").Text(),
				Url:   tmpUrl,
				Group: tmpGroup,
			})
		})

	})

	return info
}

func (x *XKMovie) nnVideoSource(sid, vid string) model.Video {
	var video = model.Video{Id: sid, Vid: vid, Source: sid}

	b, err := x.httpWrapper.Get(fmt.Sprintf(xkPlayUrl, sid))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return video
	}

	var respHtml = string(b)
	var cfUrl = fmt.Sprintf(cloudflareUrl, url.QueryEscape(fmt.Sprintf(xkPlayUrl, sid)), url.QueryEscape(".stui-content__desc"))
	respHtml = fuckCloudflare(respHtml, cfUrl)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(respHtml))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return video
	}

	video.Name, _ = doc.Find(".stui-content__desc").Find(".pic").Attr("title")
	video.Thumb, _ = doc.Find(".stui-content__desc").Find(".img-responsive").Attr("src")

	var findJson = util.SimpleRegEx(string(b), `player_aaaa=(\S+)</script>`)
	var result = gjson.Parse(findJson)
	video.Url = result.Get("url").String()
	video.Source = result.Get("url").String()
	video.Type = util.GuessVideoType(video.Url)

	return video
}
