package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
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

	ctx.HTML(http.StatusOK, "home/home.html", gin.H{
		"data": movieListByTag("zuixindianying", page),
	})
}

func (p ResourceController) Info(ctx *gin.Context) {
	var id = ctx.Param("id")

	var d = movieInfoById(id)
	ctx.JSON(http.StatusOK, d)
	return

	//ctx.HTML(http.StatusOK, "home/home.html", gin.H{
	//	"data": movieListByTag("zuixindianying", d),
	//})
}

func (p ResourceController) Video(ctx *gin.Context) {
	var id = ctx.Param("id")

	var d = movieVideoById(id)
	ctx.JSON(http.StatusOK, d)
	return
}
