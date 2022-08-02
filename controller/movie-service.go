package controller

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/ShotTv-api/model"
	"log"
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
		fmt.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf("https://www.czspp.com/xssearch?q=%s&p=%s", query, page))
	if err != nil {
		fmt.Println("[ERR]", err.Error())
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
		fmt.Println("Visiting", request.URL.String())
	})

	log.Println(fmt.Sprintf("https://www.czspp.com/%s/page/%d", tagName, _page))

	err := c.Visit(fmt.Sprintf("https://www.czspp.com/%s/page/%d", tagName, _page))
	if err != nil {
		fmt.Println("[ERR]", err.Error())
	}

	return pager
}

func movieInfoById(id string) model.MovieInfo {
	var info = model.MovieInfo{}

	c := colly.NewCollector()

	c.OnHTML(".paly_list_btn", func(element *colly.HTMLElement) {
		element.ForEach("a", func(i int, element *colly.HTMLElement) {
			info.Links = append(info.Links, model.Link{
				Name: element.Text,
				Url:  element.Attr("href"),
			})
		})
	})

	c.OnHTML(".dyxingq", func(element *colly.HTMLElement) {
		info.Thumb = element.ChildAttr(".dyimg img", "src")
		info.Name = element.ChildText(".moviedteail_tt h1")
		info.Intro = element.ChildText(".yp_context")
	})

	c.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf("https://www.czspp.com/movie/%s.html", id))
	if err != nil {
		fmt.Println("[ERR]", err.Error())
	}

	return info
}
