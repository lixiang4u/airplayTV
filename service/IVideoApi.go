package service

import (
	"github.com/gocolly/colly"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"log"
	"net/http"
	"net/url"
	"os"
)

// 定义视频源操作方法，支持多源站接入
type IVideoApi interface {
	ListByTag(tagName, page string) model.Pager
	Search(search, page string) model.Pager
	Detail(id string) model.MovieInfo
	Source(sid, vid string) model.Video
}

func HandleSrcM3U8FileToLocal(id, sourceUrl string, isCache bool) string {
	log.Println("[HandleSrcM3U8FileToLocal]", id, sourceUrl)
	tmpUrl, err := url.Parse(sourceUrl)

	// hd.njeere.com 视频文件加密了，不能直接下载m3u8文件到本地服务器

	// 如果是存在CORS切文件直接redirect后能播放的，使用如下处理
	if util.StringInList(util.HandleHost(sourceUrl), util.RedirectConfig) {
		return util.HandleUrlToCORS(sourceUrl)
	}

	if err == nil && util.StringInList(tmpUrl.Hostname(), []string{
		"hd.njeere.com",
		"s1.czspp.com",
		"yun.m3.c-zzy.online", // 这个需要特殊处理，返回的是m3u8数据，但是后缀是mp4
	}) {
		return sourceUrl
	}

	var localFile = util.NewLocalVideoFileName(id, sourceUrl)
	err = downloadSourceFile(id, sourceUrl, localFile, isCache)
	if err != nil {
		log.Println("[download.m3u8.error]", err)
		return sourceUrl
	}
	return util.GetLocalVideoFileUrl(localFile)
}

func downloadSourceFile(id, url, local string, isCache bool) (err error) {
	if util.PathExist(local) {
		return
	}

	c := Movie{IsCache: isCache}.NewColly()

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	c.OnResponse(func(response *colly.Response) {
		var f *os.File
		if response.StatusCode == http.StatusOK {
			bs := util.HandleM3U8Contents(response.Body, util.HandleHost(url))
			f, err = os.OpenFile(local, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
			_, err = f.Write(bs)
		} else {
			log.Println("[request.error]", response.StatusCode)
		}
	})

	err = c.Visit(url)
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return err
}
