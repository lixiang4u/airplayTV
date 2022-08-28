package service

import (
	"bytes"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/grafov/m3u8"
	"github.com/lixiang4u/ShotTv-api/model"
	"github.com/lixiang4u/ShotTv-api/util"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	// https://s1.czspp.com:7721 视频可以直接播放
	if err == nil && util.StringInList(tmpUrl.Hostname(), []string{"hd.njeere.com", "s1.czspp.com"}) {
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
			bs := handleEXTM3UHost(response.Body, util.HandleHost(url))
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

// 处理m3u8内容（修正地址问题）
func handleEXTM3UHost(data []byte, host string) []byte {
	if host == "" {
		return data
	}
	playList, listType, err := m3u8.DecodeFrom(bytes.NewBuffer(data), true)
	if err != nil {
		log.Println("[m3u8.DecodeFrom.error]", err)
		return data
	}

	switch listType {
	case m3u8.MEDIA:
		mediapl := playList.(*m3u8.MediaPlaylist)
		for idx, val := range mediapl.Segments {
			if val == nil {
				continue
			}
			if util.IsHttpUrl(val.URI) == false {
				mediapl.Segments[idx].URI = fmt.Sprintf("%s/%s", host, strings.TrimLeft(val.URI, "/"))
			}
		}
	case m3u8.MASTER:
		masterpl := playList.(*m3u8.MasterPlaylist)
		for idx, val := range masterpl.Variants {
			if val == nil {
				continue
			}
			if util.IsHttpUrl(val.URI) == false {
				masterpl.Variants[idx].URI = fmt.Sprintf("%s/%s", host, strings.TrimLeft(val.URI, "/"))
			}
		}
	}

	return playList.Encode().Bytes()
}
