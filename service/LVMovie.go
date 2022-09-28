package service

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"log"
	"strings"
)

var (
	lvSearchUrl = "https://lvv2.com/mv/search/q-%s/type-/page-%d" //https://lvv2.com/mv/search/q-天/type-/page-2
)

type LVMovie struct{ Movie }

func (x LVMovie) ListByTag(tagName, page string) model.Pager {
	tagName = "天"
	return x.lvListBySearch(tagName, page)
}

func (x LVMovie) Search(search, page string) model.Pager {
	return x.lvListBySearch(search, page)
}

func (x LVMovie) Detail(id string) model.MovieInfo {
	return model.MovieInfo{}
}

func (x LVMovie) Source(sid, vid string) model.Video {
	return model.Video{}
}

// ===

func (x LVMovie) lvHandleUrlToId(tmpUrl string) int {
	tmpList := strings.Split(tmpUrl, "/")
	return util.HandlePageNumber(strings.TrimLeft(tmpList[len(tmpList)-1], "page-"))
}

func (x LVMovie) lvListBySearch(query, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 24
	pager.Current = util.HandlePageNumber(page)

	c := x.Movie.NewColly()

	c.OnHTML(".wrap-content .stui-pannel-box", func(element *colly.HTMLElement) {
		name := element.ChildText(".stui-content__detail .title a")
		url := element.ChildAttr(".stui-content__detail .title a", "href")
		thumb := element.ChildAttr(".stui-content__thumb img", "data-original")
		tag := ""
		actors := ""

		pager.List = append(pager.List, model.MovieInfo{
			Id:     util.CZHandleUrlToId(url),
			Name:   name,
			Thumb:  thumb,
			Url:    url,
			Actors: actors,
			Tag:    tag,
		})
	})

	c.OnHTML("#page", func(element *colly.HTMLElement) {
		element.ForEach("a", func(i int, element *colly.HTMLElement) {
			pager.Total = pager.Limit * x.lvHandleUrlToId(element.Attr("href"))
		})
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf(lvSearchUrl, query, pager.Current))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return pager
}
