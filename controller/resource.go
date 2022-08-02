package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/ShotTv-api/model"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type ResourceController struct {
	Tag string
}

func (p ResourceController) Search(ctx *gin.Context) {
	var query = ctx.Query("q")
	var page = ctx.Query("p")

	var pager = model.Pager{}
	pager.Limit = 20

	c := colly.NewCollector()

	c.OnHTML(".search_list ul li", func(element *colly.HTMLElement) {
		name := element.ChildText(".dytit a")
		url := element.ChildAttr(".dytit a", "href")
		thumb := element.ChildAttr("img.thumb", "data-original")
		tag := element.ChildText(".nostag")
		actors := element.ChildText(".inzhuy")

		//fmt.Printf("Link found: %q %s -> %s, %s, %s \n", thumb[:10], url, name, tag, actors[:10])

		pager.List = append(pager.List, model.SearchList{
			Id:     "",
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

	ctx.JSON(http.StatusOK, pager)
}

// 根据标签获取视频列表
func (p ResourceController) ListByTag(ctx *gin.Context) {
	var tagName = ctx.Param("tagName")
	page, _ := strconv.Atoi(ctx.Query("p"))

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

		pager.List = append(pager.List, model.SearchList{
			Id:         "",
			Name:       name,
			Thumb:      thumb,
			Url:        url,
			Actors:     actors,
			Tag:        tag,
			Resolution: resolution,
		})
	})

	c.OnHTML(".pagenavi_txt", func(element *colly.HTMLElement) {
		fmt.Println("[elTxt]", element.Text)
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

	log.Println(fmt.Sprintf("https://www.czspp.com/%s/page/%d", tagName, page))

	err := c.Visit(fmt.Sprintf("https://www.czspp.com/%s/page/%d", tagName, page))
	if err != nil {
		fmt.Println("[ERR]", err.Error())
	}

	ctx.JSON(http.StatusOK, pager)
}
