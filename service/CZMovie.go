package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/ShotTv-api/model"
	"github.com/lixiang4u/ShotTv-api/util"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var (
	czTagUrl    = "https://www.czspp.com/%s/page/%d"
	czSearchUrl = "https://www.czspp.com/xssearch?q=%s&p=%s"
	czDetailUrl = "https://www.czspp.com/movie/%s.html"
	czPlayUrl   = "https://www.czspp.com/v_play/%s.html"
)

//========================================================================
//==============================接口实现===================================
//========================================================================

type CZMovie struct{}

func (x CZMovie) ListByTag(tagName, page string) model.Pager {
	return czListByTag(tagName, page)
}

func (x CZMovie) Search(search, page string) model.Pager {
	return czListBySearch(search, page)
}

func (x CZMovie) Detail(id string) model.MovieInfo {
	return czVideoDetail(id)
}

func (x CZMovie) Source(sid, vid string) model.Video {
	return czVideoSource(sid, vid)
}

//========================================================================
//==============================实际业务处理逻辑============================
//========================================================================

func czListByTag(tagName, page string) model.Pager {
	_page, _ := strconv.Atoi(page)

	var pager = model.Pager{}
	pager.Limit = 25

	c := colly.NewCollector(colly.CacheDir(util.GetCollyCacheDir()))

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

func czListBySearch(query, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 20

	c := colly.NewCollector(colly.CacheDir(util.GetCollyCacheDir()))

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

	err := c.Visit(fmt.Sprintf(czSearchUrl, query, page))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return pager
}

func czVideoDetail(id string) model.MovieInfo {
	var info = model.MovieInfo{}

	c := colly.NewCollector()

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

func czVideoSource(sid, vid string) model.Video {
	var video = model.Video{}
	var err error

	c := colly.NewCollector()

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
			video, err = czParseVideoSource(sid, findLine)

			bs, _ := json.MarshalIndent(video, "", "\t")
			log.Println(fmt.Sprintf("[video] %s", string(bs)))

			if err != nil {
				log.Println("[parse.video.error]", err)
			}
		}
	})

	c.OnHTML(".jujiinfo", func(element *colly.HTMLElement) {
		video.Name = element.ChildText("h3")
	})

	err = c.Visit(fmt.Sprintf(czPlayUrl, sid))
	if err != nil {
		log.Println("[ERR]", err.Error())
	}

	return video
}

func czParseVideoSource(id, js string) (model.Video, error) {
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

	if !strings.Contains(video.Source, "aliyundrive.asia") {
		return video, nil
	}

	video.Url = HandleSrcM3U8FileToLocal(id, video.Source)

	return video, nil
}
