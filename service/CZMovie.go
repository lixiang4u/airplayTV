package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	czTagUrl    = "https://www.czspp.com/%s/page/%d"
	czSearchUrl = "https://www.czspp.com/xssearch?q=%s&p=%d"
	czDetailUrl = "https://www.czspp.com/movie/%s.html"
	czPlayUrl   = "https://www.czspp.com/v_play/%s.html"
)

//========================================================================
//==============================接口实现===================================
//========================================================================

type CZMovie struct{ Movie }

func (x CZMovie) ListByTag(tagName, page string) model.Pager {
	return x.czListByTag(tagName, page)
}

func (x CZMovie) Search(search, page string) model.Pager {
	return x.czListBySearch(search, page)
}

func (x CZMovie) Detail(id string) model.MovieInfo {
	return x.czVideoDetail(id)
}

func (x CZMovie) Source(sid, vid string) model.Video {
	return x.czVideoSource(sid, vid)
}

//========================================================================
//==============================实际业务处理逻辑============================
//========================================================================

func (x CZMovie) czListByTag(tagName, page string) model.Pager {
	_page, _ := strconv.Atoi(page)

	var pager = model.Pager{}
	pager.Limit = 25

	c := x.Movie.NewColly()

	c.OnHTML(".mi_cont .mi_ne_kd ul li", func(element *colly.HTMLElement) {
		name := element.ChildText(".dytit a")
		url := element.ChildAttr(".dytit a", "href")
		thumb := element.ChildAttr("img.thumb", "data-original")
		tag := element.ChildText(".nostag")
		actors := element.ChildText(".inzhuy")
		resolution := element.ChildText(".hdinfo span")

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

	c.OnHTML(".pagenavi_txt", func(element *colly.HTMLElement) {
		element.ForEach("a", func(i int, element *colly.HTMLElement) {
			tmpList := strings.Split(element.Attr("href"), "/")
			n, _ := strconv.Atoi(tmpList[len(tmpList)-1])
			if n*pager.Limit > pager.Total {
				pager.Total = n * pager.Limit
			}
		})
	})

	c.OnHTML(".pagenavi_txt .current", func(element *colly.HTMLElement) {
		pager.Current, _ = strconv.Atoi(element.Text)
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	log.Println(fmt.Sprintf(czTagUrl, tagName, _page))

	err := c.Visit(fmt.Sprintf(czTagUrl, tagName, _page))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return pager
}

func (x CZMovie) czListBySearch(query, page string) model.Pager {
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

	err := c.Visit(fmt.Sprintf(czSearchUrl, query, util.HandlePageNumber(page)))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return pager
}

func (x CZMovie) czVideoDetail(id string) model.MovieInfo {
	var info = model.MovieInfo{}

	c := x.Movie.NewColly()

	c.OnHTML(".paly_list_btn", func(element *colly.HTMLElement) {
		element.ForEach("a", func(i int, element *colly.HTMLElement) {
			info.Links = append(info.Links, model.Link{
				Id:   util.CZHandleUrlToId2(element.Attr("href")),
				Name: strings.ReplaceAll(element.Text, "厂长", ""),
				Url:  element.Attr("href"),
			})
		})
	})

	c.OnHTML(".dyxingq", func(element *colly.HTMLElement) {
		info.Thumb = element.ChildAttr(".dyimg img", "src")
		info.Name = element.ChildText(".moviedteail_tt h1")
	})

	c.OnHTML(".yp_context", func(element *colly.HTMLElement) {
		info.Intro = strings.TrimSpace(element.Text)
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf(czDetailUrl, id))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return info
}

func (x CZMovie) czVideoSource(sid, vid string) model.Video {
	var video = model.Video{}
	var err error

	c := x.Movie.NewColly()

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	c.OnResponse(func(response *colly.Response) {
		var findLine = ""
		tmpList := strings.Split(string(response.Body), "\n")
		for _, line := range tmpList {
			if strings.Contains(line, "md5.AES.decrypt") {
				findLine = line
				break
			}
		}
		if findLine != "" {
			video, err = x.czParseVideoSource(sid, findLine)

			bs, _ := json.MarshalIndent(video, "", "\t")
			log.Println(fmt.Sprintf("[video] %s", string(bs)))

			if err != nil {
				log.Println("[parse.video.error]", err)
			}
		}
	})

	// 解析另一种iframe嵌套的视频
	c.OnHTML(".videoplay .viframe", func(element *colly.HTMLElement) {
		iframeUrl := element.Attr("src")
		log.Println("======[iframeUrl] ", iframeUrl)

		video.Id = sid
		video.Source, video.Type = getFrameUrlContents(iframeUrl)
		video.Url = HandleSrcM3U8FileToLocal(video.Id, video.Source, x.Movie.IsCache)
		// 1、转为本地m3u8
		// 2、修改m3u8文件内容地址,支持跨域

	})

	c.OnHTML(".jujiinfo", func(element *colly.HTMLElement) {
		video.Name = element.ChildText("h3")
	})

	err = c.Visit(fmt.Sprintf(czPlayUrl, sid))
	if err != nil {
		log.Println("[ERR]", err.Error())
	}

	// 视频类型问题处理
	video = handleVideoType(video)

	return video
}

func (x CZMovie) czParseVideoSource(id, js string) (model.Video, error) {
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

func handleVideoType(v model.Video) model.Video {
	// "https://yun.m3.c-zzy.online:1011/d/%E9%98%BF%E9%87%8C1%E5%8F%B7/%E8%8B%8F%E9%87%8C%E5%8D%97/Narco-Saints.S01E01.mp4?type=m3u8"

	tmpUrl, err := url.Parse(v.Source)
	if err != nil {
		return v
	}
	if util.StringInList(tmpUrl.Host, []string{"yun.m3.c-zzy.online:1011"}) {
		v.Type = "hls"
	}
	return v
}

func getFrameUrlContents(frameUrl string) (sourceUrl, videoType string) {
	sourceUrl = frameUrl
	videoType = "auto"

	resp, err := http.Get(frameUrl)
	if err != nil {
		log.Println("[getFrameUrlContents.get.error]", err.Error())
		return
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("[getFrameUrlContents.body.error]", err.Error())
		return
	}

	// 匹配播放文件
	regEx := regexp.MustCompile(`sources: \[{(\s+)src: '(\S+)',(\s+)type: '(\S+)'`)
	r := regEx.FindStringSubmatch(string(bs))
	if len(r) < 4 {
		return
	}
	sourceUrl = r[2]

	switch r[4] {
	case "application/vnd.apple.mpegurl":
		videoType = "hls"
	}

	return
}
