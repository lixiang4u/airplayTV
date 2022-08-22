package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lixiang4u/ShotTv-api/service"
	"net/http"
)

type VideoController struct {
	instant service.IVideoApi
}

func (x VideoController) getInstance(ctx *gin.Context) service.IVideoApi {
	var source = ctx.Query("_source")
	switch source {
	case "cz":
		x.instant = service.CZMovie{}
		break
	case "nn":
		x.instant = service.NNMovie{}
		break
	case "91":
		x.instant = service.MYMovie{}
		break
	//case "88":// 资源不行，有些都不能播放？？？
	//	x.instant = service.EightMovie{}
	//	break
	default:
		x.instant = service.NNMovie{}
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
