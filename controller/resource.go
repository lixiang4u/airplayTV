package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/ShotTv-api/model"
	"net/http"
	"strconv"
)

type ResourceController struct {
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

	// https://www.czspp.com/xssearch?q=%E5%A4%A9
	err := c.Visit(fmt.Sprintf("https://www.czspp.com/xssearch?q=%s&p=%s", query, page))
	if err != nil {
		fmt.Println("[ERR]", err.Error())
	}

	// https://www.czspp.com/xssearch?q=%E5%A4%A9
	// https://www.czspp.com/xssearch?q=天
	// 直接搜索形式展示，不用默认列表展示

	ctx.JSON(http.StatusOK, pager)
}
