package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/ShotTv-api/model"
	"github.com/lixiang4u/ShotTv-api/util"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func movieListBySearch(query, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 20

	c := colly.NewCollector()

	c.OnHTML(".search_list ul li", func(element *colly.HTMLElement) {
		name := element.ChildText(".dytit a")
		url := element.ChildAttr(".dytit a", "href")
		thumb := element.ChildAttr("img.thumb", "data-original")
		tag := element.ChildText(".nostag")
		actors := element.ChildText(".inzhuy")

		pager.List = append(pager.List, model.MovieInfo{
			Id:     handleUrlToId(url),
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

	err := c.Visit(fmt.Sprintf("https://www.czspp.com/xssearch?q=%s&p=%s", query, page))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return pager
}

func movieListByTag(tagName, page string) model.Pager {
	_page, _ := strconv.Atoi(page)

	var pager = model.Pager{}
	pager.Limit = 25

	c := colly.NewCollector()

	c.OnHTML(".mi_cont .mi_ne_kd ul li", func(element *colly.HTMLElement) {
		name := element.ChildText(".dytit a")
		url := element.ChildAttr(".dytit a", "href")
		thumb := element.ChildAttr("img.thumb", "data-original")
		tag := element.ChildText(".nostag")
		actors := element.ChildText(".inzhuy")
		resolution := element.ChildText(".hdinfo span")

		pager.List = append(pager.List, model.MovieInfo{
			Id:         handleUrlToId(url),
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

	log.Println(fmt.Sprintf("https://www.czspp.com/%s/page/%d", tagName, _page))

	err := c.Visit(fmt.Sprintf("https://www.czspp.com/%s/page/%d", tagName, _page))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return pager
}

func movieInfoById(id string) model.MovieInfo {
	var info = model.MovieInfo{}

	c := colly.NewCollector()

	c.OnHTML(".paly_list_btn", func(element *colly.HTMLElement) {
		element.ForEach("a", func(i int, element *colly.HTMLElement) {
			info.Links = append(info.Links, model.Link{
				Id:   handleUrlToId2(element.Attr("href")),
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

	err := c.Visit(fmt.Sprintf("https://www.czspp.com/movie/%s.html", id))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return info
}

func movieVideoById(id string) model.Video {
	var video = model.Video{}
	var err error

	c := colly.NewCollector()

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	c.OnResponse(func(response *colly.Response) {
		var findLine = ""
		tmplist := strings.Split(string(response.Body), "\n")
		for _, line := range tmplist {
			if strings.Contains(line, "md5.AES.decrypt") {
				findLine = line
				break
			}
		}
		if findLine != "" {
			video, err = parseVideo(id, findLine)

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

	err = c.Visit(fmt.Sprintf("https://www.czspp.com/v_play/%s.html", id))
	if err != nil {
		log.Println("[ERR]", err.Error())
	}

	return video
}

func parseVideo(id, js string) (model.Video, error) {
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

	var localFile = util.NewLocalVideoFileName(id, video.Source)

	if !strings.Contains(video.Source, "aliyundrive.asia") {
		return video, nil
	}

	err = downloadFile(id, video.Source, localFile)
	if err != nil {
		return video, err
	}
	video.Url = util.GetLocalVideoFileUrl(localFile)

	return video, nil
}

func downloadFile(id, url, local string) (err error) {
	if util.PathExist(local) {
		return
	}

	c := colly.NewCollector()

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	c.OnResponse(func(response *colly.Response) {
		var f *os.File
		if response.StatusCode == http.StatusOK {
			f, err = os.OpenFile(local, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
			_, err = f.Write(response.Body)
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
