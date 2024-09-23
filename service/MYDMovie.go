package service

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"github.com/zc310/headers"
	"log"
	"strings"
)

var (
	mydHost      = "https://myd04.com/"
	mydImageHost = "https://www.mdzypic.com/"
	mydTagUrl    = "https://myd04.com/vodshow/1--------%d---.html"
	mydM3u8Url   = "https://nnyy.in/url.php"
	mydPlayUrl   = "https://www.huibangpaint.com/vodplay/%s.html"
	mydSearchUrl = "https://myd04.com/vodsearch/%s----------%d---.html"
	mydDetailUrl = "https://www.huibangpaint.com/voddetail/%s.html"
)

type MYDMovie struct {
	movie       Movie
	httpWrapper *util.HttpWrapper
}

func (x *MYDMovie) Init(movie Movie) {
	x.movie = movie
	if x.httpWrapper == nil {
		x.httpWrapper = &util.HttpWrapper{}
	}

	x.httpWrapper.SetHeader(headers.Origin, xkHost)
	x.httpWrapper.SetHeader(headers.Referer, xkHost)
	x.httpWrapper.SetHeader(headers.UserAgent, ua)
	x.httpWrapper.SetHeader(headers.UserAgent, ua)
}

func (x *MYDMovie) ListByTag(tagName, page string) model.Pager {
	return x._ListByTag("", page)
}

func (x *MYDMovie) Search(search, page string) model.Pager {
	return x._ListBySearch(search, page)
}

func (x *MYDMovie) Detail(id string) model.MovieInfo {
	return x._VideoDetail(id)
}

func (x *MYDMovie) Source(sid, vid string) model.Video {
	return x._VideoSource(sid, vid)
}

//========================================================================
//==============================实际业务处理逻辑============================
//========================================================================

func (x *MYDMovie) _ListBySearch(search, page string) model.Pager {
	var pager = model.Pager{Limit: 15, Current: util.HandlePageNumber(page)}

	b, err := x.httpWrapper.Get(fmt.Sprintf(mydSearchUrl, search, pager.Current))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return pager
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return pager
	}

	doc.Find(".module .module-item").Each(func(i int, selection *goquery.Selection) {
		name := selection.Find(".module-card-item-title").Text()
		tmpUrl, _ := selection.Find(".module-card-item-title a").Attr("href")
		thumb, _ := selection.Find(".lazyload").Attr("data-original")
		tag := selection.Find(".module-item-note").Text()
		intro := selection.Find(".module-info-item-content").Text()

		pager.List = append(pager.List, model.MovieInfo{
			Id:         util.SimpleRegEx(tmpUrl, `(\d+)`),
			Name:       strings.TrimSpace(name),
			Thumb:      util.FillUrlHost(thumb, mydImageHost),
			Url:        util.FillUrlHost(tmpUrl, mydImageHost),
			Tag:        tag,
			Resolution: tag,
			Intro:      strings.TrimSpace(intro),
		})
	})

	return pager
}

func (x *MYDMovie) _ListByTag(tagName, page string) model.Pager {
	var pager = model.Pager{Limit: 72, Current: util.HandlePageNumber(page)}

	b, err := x.httpWrapper.Get(fmt.Sprintf(mydTagUrl, pager.Current))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return pager
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return pager
	}

	doc.Find(".module .module-item").Each(func(i int, selection *goquery.Selection) {
		name := selection.Find(".module-poster-item-title").Text()
		tmpUrl, _ := selection.Attr("href")
		thumb, _ := selection.Find(".lazyload").Attr("data-original")
		tag := selection.Find(".module-item-note").Text()

		pager.List = append(pager.List, model.MovieInfo{
			Id:         util.SimpleRegEx(tmpUrl, `(\d+)`),
			Name:       name,
			Thumb:      util.FillUrlHost(thumb, mydImageHost),
			Url:        util.FillUrlHost(tmpUrl, mydImageHost),
			Tag:        tag,
			Resolution: tag,
		})
	})

	return pager
}

// 根据id获取视频播放列表信息
func (x *MYDMovie) _VideoDetail(id string) model.MovieInfo {
	var info = model.MovieInfo{Id: id}
	//var sourceMap = make(map[string]string, 0)

	//c := x.Movie.NewColly()
	//c.OnHTML(".myui-content__thumb", func(element *colly.HTMLElement) {
	//	info.Thumb = util.FillUrlHost(element.ChildAttr("a img", "data-original"), mydHost)
	//	info.Name = element.ChildAttr("a", "title")
	//})
	//c.OnHTML("meta[name=description]", func(element *colly.HTMLElement) {
	//	info.Intro = element.Attr("content")
	//})
	//c.OnHTML(".myui-panel_hd .nav-tabs", func(element *colly.HTMLElement) {
	//	element.ForEach("li a", func(i int, element *colly.HTMLElement) {
	//		sourceMap[strings.TrimLeft(element.Attr("href"), "#")] = element.Text
	//	})
	//})
	//c.OnHTML(".tab-content", func(element *colly.HTMLElement) {
	//	element.ForEach(".tab-pane", func(groupIndex int, element *colly.HTMLElement) {
	//		var sourceId = element.Attr("id")
	//		groupName, ok := sourceMap[sourceId]
	//		if !ok {
	//			groupName = fmt.Sprintf("来源%d", groupIndex+1)
	//		}
	//		element.ForEach("li a", func(i int, element *colly.HTMLElement) {
	//			info.Links = append(info.Links, model.Link{
	//				Id:    util.SimpleRegEx(element.Attr("href"), `(\d+-\d+-\d+)`),
	//				Name:  element.Text,
	//				Url:   util.FillUrlHost(element.Attr("href"), mydHost),
	//				Group: groupName,
	//			})
	//		})
	//	})
	//})
	//
	//c.OnRequest(func(request *colly.Request) {
	//	log.Println("Visiting", request.URL.String())
	//})
	//
	//err := c.Visit(fmt.Sprintf(mydDetailUrl, id))
	//if err != nil {
	//	log.Println("[visit.error]", err.Error())
	//}

	return info
}

func (x *MYDMovie) _VideoSource(sid, vid string) model.Video {
	var video = model.Video{Id: sid, Source: sid}

	//获取基础信息
	//c := x.Movie.NewColly()
	//c.OnHTML(".myui-player__data", func(element *colly.HTMLElement) {
	//	video.Name = element.ChildText(".text-fff")
	//	video.Thumb = ""
	//})
	//c.OnHTML(".embed-responsive", func(element *colly.HTMLElement) {
	//	video.Source = util.SimpleRegEx(element.Text, `"url":"(\S+?)",`)
	//	video.Source = strings.ReplaceAll(video.Source, "\\/", "/")
	//	video.Type = util.GuessVideoType(video.Source)
	//
	//	video.Url = HandleSrcM3U8FileToLocal(sid, video.Source, x.Movie.IsCache)
	//})
	//
	//c.OnRequest(func(request *colly.Request) {
	//	log.Println("Visiting", request.URL.String())
	//})
	//
	//err := c.Visit(fmt.Sprintf(mydPlayUrl, sid))
	//if err != nil {
	//	log.Println("[visit.error]", err.Error())
	//}

	return video
}
