package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lixiang4u/airplayTV/service"
	"github.com/lixiang4u/airplayTV/util"
	go_websocket "github.com/lixiang4u/go-websocket"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type VideoController struct {
	instant service.IVideoApi
}

// 解析缓存变量
func handleCache(cacheStr string) bool {
	cacheStr = strings.ToLower(cacheStr)
	if cacheStr == "" { // 没设置cache默认为缓存
		return false
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
		_m := service.CZMovie{}
		_m.Init(m)
		x.instant = &_m
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
	case "lv":
		x.instant = service.LVMovie{Movie: m}
		break
	case "five":
		_m := service.FiveMovie{}
		_m.Init(m)
		x.instant = &_m
		break
	case "yanf":
		_m := service.YaNF{}
		_m.Init(m)
		x.instant = &_m
		break
	//case "88":// 资源不行，有些都不能播放？？？
	//	x.instant = service.EightMovie{}
	//	break
	case "ys":
		_m := service.YSMovie{}
		_m.Init(m)
		x.instant = &_m
		break
	case "xk":
		_m := service.XKMovie{}
		_m.Init(m)
		x.instant = &_m
		break
	default:
		_m := service.YSMovie{}
		_m.Init(m)
		x.instant = &_m
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
	var id = ctx.Query("id")        // 播放id
	var vid = ctx.Query("vid")      // 视频id
	var m3u8p = ctx.Query("_m3u8p") // 视频id

	var data = x.getInstance(ctx).Source(id, vid)

	if m3u8p == "true" {
		data.Url = fmt.Sprintf("https://%s/api/m3u8p?q=%s", ctx.Request.Host, data.Url)
	}

	ctx.JSON(http.StatusOK, data)
}

// 隔空播放，发websocket消息
func (x VideoController) Airplay(ctx *gin.Context) {
	var id = ctx.Query("id")              // 播放id
	var vid = ctx.Query("vid")            // 视频id
	var clientId = ctx.Query("client_id") // 客户端id
	var source = ctx.Query("_source")     // 投射需要该参数
	var m3u8p = ctx.Query("_m3u8p")       //

	var d = gin.H{
		"event":     "play",
		"client_id": clientId,
		"_source":   source,
		"_m3u8p":    m3u8p,
		"video":     x.getInstance(ctx).Source(id, vid),
		"timestamp": time.Now().Unix(),
	}
	b, _ := json.MarshalIndent(d, "", "\t")

	log.Println("[debug]", clientId, string(b))

	if go_websocket.WSendMessage(clientId, websocket.TextMessage, b) == false {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "msg": "发送失败或TV不在线", "data": nil})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "发送成功", "data": nil})
}

func (x VideoController) VideoControls(ctx *gin.Context) {
	var clientId = ctx.Query("client_id") // 客户端id
	var control = ctx.Query("control")    // 投射需要该参数
	var value = ctx.Query("value")        // 投射需要该参数值

	var d = gin.H{
		"event":     "video_controls",
		"client_id": clientId,
		"control":   control,
		"value":     value,
		"timestamp": time.Now().Unix(),
	}
	b, _ := json.MarshalIndent(d, "", "\t")

	log.Println("[debug]", clientId, string(b))

	if go_websocket.WSendMessage(clientId, websocket.TextMessage, b) == false {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "msg": "发送失败或TV不在线", "data": nil})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "发送成功", "data": nil})
}

func (x VideoController) VideoVideoCORS(ctx *gin.Context) {
	var src = ctx.Query("src")
	var redirect = util.HandleRedirectUrl(src)
	log.Println("====>src: ", src)
	log.Println("====>redirect: ", redirect)
	ctx.Redirect(http.StatusTemporaryRedirect, redirect)
}
