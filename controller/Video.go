package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lixiang4u/ShotTv-api/service"
	"net/http"
	"strconv"
	"strings"
)

type VideoController struct {
	instant service.IVideoApi
}

// 解析缓存变量
func handleCache(cacheStr string) bool {
	cacheStr = strings.ToLower(cacheStr)
	if cacheStr == "" { // 没设置cache默认为缓存
		return true
	}
	// 兼容前段配置值
	switch cacheStr {
	case "open":
		return true
	case "close":
		return false
	}

	isCache, err := strconv.ParseBool(cacheStr)
	if err != nil { // 如果解析错误则默认为缓存
		return true
	}
	//只有明确指定不缓存才会不缓存
	return isCache
}

func (x VideoController) getInstance(ctx *gin.Context) service.IVideoApi {
	var source = ctx.Query("_source")
	var tmpCache = ctx.Query("_cache")

	var m = service.Movie{IsCache: handleCache(tmpCache)}

	switch source {
	case "cz":
		x.instant = service.CZMovie{Movie: m}
		break
	case "nn":
		x.instant = service.NNMovie{Movie: m}
		break
	case "91":
		x.instant = service.MYMovie{Movie: m}
		break
	case "my":
		x.instant = service.MYMovie{Movie: m}
		break
	//case "88":// 资源不行，有些都不能播放？？？
	//	x.instant = service.EightMovie{}
	//	break
	default:
		x.instant = service.CZMovie{Movie: m}
		//ctx.JSON(http.StatusOK, gin.H{"msg": "source not exists"})
	}
	return x.instant
}

func (x VideoController) ListByTag(ctx *gin.Context) {
	var tagName = ctx.Query("tagName")
	var page = ctx.Query("p")

	var data = x.getInstance(ctx).ListByTag(tagName, page)

	ctx.JSON(http.StatusOK, data)
}

func (x VideoController) Search(ctx *gin.Context) {
	var query = ctx.Query("q")
	var page = ctx.Query("p")

	var data = x.getInstance(ctx).Search(query, page)

	ctx.JSON(http.StatusOK, data)
}

func (x VideoController) Detail(ctx *gin.Context) {
	var id = ctx.Param("id")

	var data = x.getInstance(ctx).Detail(id)

	ctx.JSON(http.StatusOK, data)
}

func (x VideoController) Source(ctx *gin.Context) {
	var id = ctx.Param("id")   // 播放id
	var vid = ctx.Query("vid") // 视频id

	var data = x.getInstance(ctx).Source(id, vid)

	ctx.JSON(http.StatusOK, data)
}

func (x VideoController) ListByTagV2(ctx *gin.Context) {
	var tagName = ctx.Query("tagName")
	var page = ctx.Query("p")

	var data = x.getInstance(ctx).ListByTag(tagName, page)

	ctx.JSON(http.StatusOK, data)
}

func (x VideoController) DetailV2(ctx *gin.Context) {
	var id = ctx.Query("id")

	var data = x.getInstance(ctx).Detail(id)

	ctx.JSON(http.StatusOK, data)
}

func (x VideoController) SourceV2(ctx *gin.Context) {
	var id = ctx.Query("id")   // 播放id
	var vid = ctx.Query("vid") // 视频id

	var data = x.getInstance(ctx).Source(id, vid)

	ctx.JSON(http.StatusOK, data)
}
