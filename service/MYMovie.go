package service

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/grafov/m3u8"
	"github.com/lixiang4u/ShotTv-api/model"
	"github.com/lixiang4u/ShotTv-api/util"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var (
	myPlayUrl   = "https://www.91mayi.com/vodplay/%s.html" //https://www.91mayi.com/vodplay/190119-1-30.html
	myParseUrl  = "https://zj.shankuwang.com:8443/?url=%s" // 云解析
	myDetailUrl = "https://www.91mayi.com/voddetail/%s.html"
	mySearchUrl = "https://www.91mayi.com/vodsearch/%s----------%d---.html"
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

	err := c.Visit(fmt.Sprintf(mySearchUrl, search, util.HandlePageNumber(page)))
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

	c.OnHTML(".col-md-wide-75", func(element *colly.HTMLElement) {
		info.Thumb = element.ChildAttr("a.v-thumb .lazyload", "data-original")
		info.Name = element.ChildAttr("a.v-thumb", "title")
		info.Intro = element.ChildText(".detail-content")
		info.Tag = element.ChildText(".pic-text")
		info.Url = element.ChildAttr("a.v-thumb", "href")
	})

	c.OnHTML(".stui-content__playlist", func(element *colly.HTMLElement) {
		element.ForEach("li a", func(i int, element *colly.HTMLElement) {
			info.Links = append(info.Links, model.Link{
				Id:   myHandlePlayUrlId(element.Attr("href")),
				Name: element.Text,
				Url:  element.Attr("href"),
			})
		})
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf(myDetailUrl, id))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return info
}

// 使用chromedp直接请求页面关联的播放数据m3u8
// 应该可以直接从chromedp拿到m3u8地址，但是没跑通，可以先拿到请求所需的所有上下文，然后http.Post拿数据
func myVideoSource(sid, vid string) model.Video {
	var video = model.Video{Id: sid, Source: sid}

	//获取基础信息
	c := colly.NewCollector(colly.CacheDir(util.GetCollyCacheDir()))

	c.OnHTML(".stui-content__thumb", func(element *colly.HTMLElement) {
		video.Name = element.ChildAttr(".v-thumb", "title")
		video.Thumb = element.ChildAttr(".v-thumb .lazyload", "src")
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf(myDetailUrl, vid))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	video.Type = "hls"
	video.Url = handleMYVideoUrl(sid)

	return video
}

func myHandlePlayUrlId(url string) (id string) {
	tmpList := strings.Split(strings.Trim(strings.Trim(url, ".html"), "/"), "/")
	if len(tmpList) == 2 {
		return tmpList[1]
	}
	return
}

func handleMYVideoUrl(id string) (tmpUrl string) {
	tmpSid := myFetchPlayInfo(id)
	tmpUrl = myCloudParse(tmpSid)
	return tmpUrl
}

// 获取播放文件流程：
//https://www.91mayi.com/vodplay/190119-1-30.html
//=>
//JS: "url": "mayi_....
//=>
//https://zj.shankuwang.com:8443/?url=mayi_cb52u6qXdjt2EuviqgoWCYpNnbRCAr3tZE9aGjb%2Fi%2BzDKmZoJSUGqTzs6B75QYq6iD4yXYfRNGn%2BSuZm30SufKP0IaBoamqFUlDtjZfhQA
//=>
//https://new.qqaku.com/20220817/YHUGLoN8/index.m3u8

func myFetchPlayInfo(id string) (tmpSid string) {
	c := colly.NewCollector()

	c.OnResponse(func(response *colly.Response) {
		// "url":"mayi_f89adGIRDSJxhFCwJcrVTunxt3eQ%2B8xZgSn8fb0QQSKVbR5zTdl0fF890A8oZpC9IYqnL5ScIxuA%2BWldL3%2Fc2Uwy48E","url_next":"mayi_ac5aTQEQLN%2BKpYBZiGiLgOqVh2cE83GT1hLc8M8","from":"iqiyi"
		regex := regexp.MustCompile(`"url":"(\S+)","url_next"`)
		matches := regex.FindStringSubmatch(string(response.Body))
		if len(matches) > 1 && strings.HasPrefix(matches[1], "mayi_") {
			tmpSid = matches[1]
		}
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf(myPlayUrl, id))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return
}

func myCloudParse(id string) (tmpUrl string) {
	c := colly.NewCollector()

	c.OnResponse(func(response *colly.Response) {
		regex := regexp.MustCompile(`var video_url = '(\S+)';`)
		matches := regex.FindStringSubmatch(string(response.Body))
		// 没匹配到
		if len(matches) <= 1 {
			return
		}
		// 是不http协议数据
		if !strings.HasPrefix(matches[1], "http") {
			return
		}
		// 返回的iqiyi错误地址可能就是html结尾的
		if strings.HasSuffix(matches[1], ".html") {
			return
		}
		tmpUrl = matches[1]
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf(myParseUrl, id))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return
}

func myCheckVideoUrlRedirect(tmpUrl string) {
	// https://new.qqaku.com/20220727/kBNdvBbi/index.m3u8

	c := colly.NewCollector()

	c.OnResponse(func(response *colly.Response) {

		playList, listType, err := m3u8.DecodeFrom(strings.NewReader(string(response.Body)), true)
		if err != nil {
			log.Println("[decode.err]", err)
			return
		}

		switch listType {
		case m3u8.MEDIA:
			mediapl := playList.(*m3u8.MediaPlaylist)
			fmt.Printf("[0x01] %+v\n", mediapl)
		case m3u8.MASTER:
			masterpl := playList.(*m3u8.MasterPlaylist)
			fmt.Printf("[0x02]%+v\n", masterpl)
			//log.Println("[]", masterpl.)
		}

		//log.Println("[redirect]", string(response.Body))
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(tmpUrl)
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return
}
