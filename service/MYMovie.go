package service

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/ShotTv-api/model"
	"github.com/lixiang4u/ShotTv-api/util"
	"log"
	"net/url"
	"strconv"
	"strings"
)

var (
	myM3u8Url   = ""
	myPlayUrl   = ""
	mySearchUrl = "https://www.91mayi.com/vodsearch/%s----------%s---.html" // https://www.91mayi.com/vodsearch/天----------917---.html
)

type MYMovie struct{}

func (x MYMovie) ListByTag(tagName, page string) model.Pager {
	return myListBySearch("天", page)
}

func (x MYMovie) Search(search, page string) model.Pager {
	return myListBySearch(search, page)
}

func (x MYMovie) Detail(id string) model.MovieInfo {
	return myVideoDetail(id)
}

func (x MYMovie) Source(sid, vid string) model.Video {
	return myVideoSource(sid, vid)
}

//========================================================================
//==============================实际业务处理逻辑============================
//========================================================================

func myListBySearch(search, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 10

	c := colly.NewCollector(colly.CacheDir(util.GetCollyCacheDir()))

	c.OnHTML(".col-lg-wide-75 .stui-vodlist__media li", func(element *colly.HTMLElement) {
		name := element.ChildText(".title a")
		url1 := element.ChildAttr(".title a", "href")
		thumb := element.ChildAttr(".v-thumb", "data-original")
		tag := element.ChildText(".pic-text")

		pager.List = append(pager.List, model.MovieInfo{
			Id:    util.CZHandleUrlToId(url1),
			Name:  name,
			Thumb: thumb,
			Url:   url1,
			Tag:   tag,
		})
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	c.OnHTML(".stui-page .num", func(element *colly.HTMLElement) {
		tmpList := strings.Split(element.Text, "/")
		if len(tmpList) != 2 {
			return
		}
		currentIndex, _ := strconv.Atoi(tmpList[0])
		totalIndex, _ := strconv.Atoi(tmpList[1])

		pager.Current = currentIndex
		pager.Total = pager.Limit * totalIndex
	})

	err := c.Visit(fmt.Sprintf(mySearchUrl, search, page))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return pager
}

// 根据id获取视频播放列表信息
func myVideoDetail(id string) model.MovieInfo {
	var info = model.MovieInfo{}

	info.Id = id

	c := colly.NewCollector(colly.CacheDir(util.GetCollyCacheDir()))

	c.OnHTML(".product-header", func(element *colly.HTMLElement) {
		info.Thumb = element.ChildAttr(".thumb", "src")
		info.Name = element.ChildText(".product-title")
		info.Intro = element.ChildText(".product-excerpt span")
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf(nnPlayUrl, id))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}
	urls, err := handleNNVideoPlayLinks(id)

	if err == nil {
		info.Links = wrapLinks(urls)
	}

	return info
}

// 使用chromedp直接请求页面关联的播放数据m3u8
// 应该可以直接从chromedp拿到m3u8地址，但是没跑通，可以先拿到请求所需的所有上下文，然后http.Post拿数据
func myVideoSource(sid, vid string) model.Video {
	var video = model.Video{Id: sid, Source: sid}

	//获取基础信息
	c := colly.NewCollector(colly.CacheDir(util.GetCollyCacheDir()))

	c.OnHTML(".product-header", func(element *colly.HTMLElement) {
		video.Name = element.ChildText(".product-title")
		video.Thumb = element.ChildAttr(".thumb", "src")
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf(nnPlayUrl, vid))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	var v = url.Values{}
	v.Add("url", sid)
	//v.Add("sign", strconv.FormatInt(time.Now().Unix(), 10))
	_ = handleNNVideoUrl(v.Encode(), &video.Source)
	video.Type = "hls" // m3u8 都是hls ???

	video.Url = HandleSrcM3U8FileToLocal(sid, video.Source)

	return video
}
