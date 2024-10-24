package service

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"github.com/zc310/headers"
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

func IsCtYunFileUrl(tmpUrl string) bool {
	// https://media-tjwq-fy-home.tj3oss.ctyunxs.cn/FAMILYCLOUD/4839bf22-70c1-4d66-b6ba-58010898decd.mp4?x-amz-CLIENTTYPEIN=PC&AWSAccessKeyId=0Lg7dAq3ZfHvePP8DKEU&x-amz-limitrate=61440&response-content-type=video/mp4&x-amz-UID=300000534202485&response-content-disposition=attachment%3Bfilename%3D%22%E5%BC%82%E5%BD%A2%E5%A4%BA%E5%91%BD%E8%88%B02024.mp4%22%3Bfilename*%3DUTF-8%27%27%25E5%25BC%2582%25E5%25BD%25A2%25E5%25A4%25BA%25E5%2591%25BD%25E8%2588%25B02024.mp4&x-amz-OPERID=300000534202632&x-amz-CLIENTNETWORK=UNKNOWN&x-amz-CLOUDTYPEIN=FAMILY&Signature=EhN9jUkXb1cjMjWpXJlAneFsoNI%3D&Expires=1729761530&x-amz-FSIZE=3767239833&x-amz-UFID=324251161015832564
	parsed, err := url.Parse(tmpUrl)
	if err != nil {
		return false
	}
	var tmpNameList = strings.Split(parsed.Host, ".")
	if len(tmpNameList) <= 1 {
		return false
	}
	var hostname = fmt.Sprintf("%s.%s", tmpNameList[len(tmpNameList)-2], tmpNameList[len(tmpNameList)-1])
	if strings.ToLower(hostname) != "ctyunxs.cn" {
		return false
	}
	return true
}

func HandleSrcM3U8FileToLocal(id, sourceUrl string, isCache bool) string {
	log.Println("[HandleSrcM3U8FileToLocal]", id, sourceUrl)
	tmpUrl, err := url.Parse(sourceUrl)

	// hd.njeere.com 视频文件加密了，不能直接下载m3u8文件到本地服务器

	// 如果是存在CORS切文件直接redirect后能播放的，使用如下处理
	if util.StringInList(util.HandleHost(sourceUrl), util.RedirectConfig) {
		return util.HandleUrlToCORS(sourceUrl)
	}

	// ctyun一般都是mp4文件
	if IsCtYunFileUrl(sourceUrl) {
		return sourceUrl
	}

	// 直接播放的地址
	if util.StringInList(util.HandleHost(sourceUrl), util.DirectConfig) {
		return sourceUrl
	}
	if err == nil && util.StringInList(tmpUrl.Hostname(), []string{
		"hd.njeere.com",
		"s1.czspp.com",
		"yun.m3.c-zzy.online", // 这个需要特殊处理，返回的是m3u8数据，但是后缀是mp4
		"vt1.doubanio.com",    //https://vt1.doubanio.com/202211281908/3009135258a3492c9c260af15c3c9027/view/movie/M/402970553.mp4
	}) {
		return sourceUrl
	}

	var httpWrapper = util.HttpWrapper{}
	header, err := httpWrapper.Head(sourceUrl)
	if err == nil && strings.ToLower(header.Get(headers.ContentType)) == "video/mp4" {
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
	log.Println("[download.local]", local)
	if util.PathExist(local) {
		return
	}

	c := Movie{IsCache: isCache}.NewColly()

	c.OnRequest(func(request *colly.Request) {
		setHttpRequestHeader(request)
		log.Println("Visiting", request.URL.String())
	})

	c.OnResponse(func(response *colly.Response) {
		var f *os.File
		if response.StatusCode == http.StatusOK {
			bs := util.HandleM3U8Contents(response.Body, url)
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

// 支持colly请求的referer设置
func setHttpRequestHeader(request *colly.Request) {
	// 支持请求referer
	tmpReqHost := util.HandleHost(request.URL.String())
	if v, ok := util.RefererConfig[tmpReqHost]; ok {
		request.Headers.Set("Referer", v)
	} else {
		request.Headers.Set("Referer", tmpReqHost)
	}
}
