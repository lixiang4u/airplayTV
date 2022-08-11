package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lixiang4u/ShotTv-api/model"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type ResourceController struct {
	Tag string
}

func (p ResourceController) Search(ctx *gin.Context) {
	var query = ctx.Query("q")
	var page = ctx.Query("p")

	var data = movieListBySearch(query, page)

	ctx.JSON(http.StatusOK, data)
}

// 根据标签获取视频列表
func (p ResourceController) ListByTag(ctx *gin.Context) {
	var tagName = ctx.Param("tagName")
	var page = ctx.Query("p")

	var result = movieListByTag(tagName, page)

	ctx.JSON(http.StatusOK, result)
}

func handleUrlToId(url string) string {
	regex := regexp.MustCompile(`(\d{1,9})`)
	return regex.FindString(url)
}

func handleUrlToId2(url string) string {
	tmpList := strings.Split(url, "/")
	return strings.Trim(tmpList[len(tmpList)-1], ".html")
}

func Info() {

}

func (p ResourceController) Home(ctx *gin.Context) {
	var page = ctx.Query("p")
	var tvId = ctx.Query("tv_id")

	if len(tvId) > 8 {
		ctx.SetCookie("tv_id", tvId, 0, "", "", false, false)
		ctx.Redirect(302, ctx.FullPath()) //HTTP重定向:301(永久)与302(临时)
	}
	tvId, _ = ctx.Cookie("tv_id")

	ctx.HTML(http.StatusOK, "home/home.html", gin.H{
		"data":  movieListByTag("zuixindianying", page),
		"tv_id": tvId,
	})
}

func (p ResourceController) Info(ctx *gin.Context) {
	var id = ctx.Param("id")
	var unbind = ctx.Query("unbind")

	if unbind == "1" {
		ctx.SetCookie("tv_id", "", 0, "", "", false, false)
		ctx.Redirect(302, ctx.FullPath()) //HTTP重定向:301(永久)与302(临时)
	}

	var d = movieInfoById(id)

	tvId, _ := ctx.Cookie("tv_id")

	ctx.HTML(http.StatusOK, "home/info.html", gin.H{
		"data":  d,
		"tv_id": tvId,
	})
}

func (p ResourceController) VideoDetail(ctx *gin.Context) {
	var id = ctx.Param("id")

	var d = movieInfoById(id)

	ctx.JSON(http.StatusOK, d)
}

func (p ResourceController) VideoSource(ctx *gin.Context) {
	var id = ctx.Param("id")

	var d = movieVideoById(id)
	ctx.JSON(http.StatusOK, d)
	return
}

// 扫码后的页面
func (p ResourceController) Home2(ctx *gin.Context) {
	var page = ctx.Query("p")
	var search = ctx.Query("q")
	var tvId = ctx.Query("tv_id")
	var unbind = ctx.Query("unbind")

	if unbind == "1" {
		ctx.SetCookie("tv_id", "", 0, "", "", false, false)
		ctx.Redirect(302, ctx.FullPath()) //HTTP重定向:301(永久)与302(临时)
	}
	if len(tvId) > 8 {
		ctx.SetCookie("tv_id", tvId, 0, "", "", false, false)
		ctx.Redirect(302, ctx.FullPath()) //HTTP重定向:301(永久)与302(临时)
	}
	tvId, _ = ctx.Cookie("tv_id")

	var data model.Pager
	if strings.TrimSpace(search) == "" {
		data = movieListByTag("zuixindianying", page)
	} else {
		data = movieListBySearch(search, page)
	}

	var pageCurrent = 1
	var pagePrev = 1
	var pageNext = 1
	pageCurrent, _ = strconv.Atoi(page)
	if pageCurrent <= 1 {
		pagePrev = 1
		pageCurrent = 1
		pageNext = 2
	} else {
		pageNext = pageCurrent + 1
	}

	ctx.HTML(http.StatusOK, "home/home.html", gin.H{
		"data":        data,
		"tv_id":       tvId,
		"page":        page,
		"pagePrev":    pagePrev,
		"pageCurrent": pageCurrent,
		"pageNext":    pageNext,
		"search":      search,
	})
}
