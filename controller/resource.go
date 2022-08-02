package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
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

func Info() {

}
