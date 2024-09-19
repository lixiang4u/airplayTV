package controller

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/grafov/m3u8"
	"github.com/lixiang4u/airplayTV/service"
	"github.com/lixiang4u/airplayTV/util"
	"github.com/zc310/headers"
	"io"
	"log"
	"net/http"
	"os"
)

type M3u8Controller struct {
	instant     service.IVideoApi
	httpWrapper *util.HttpWrapper
}

func (x *M3u8Controller) Init() {
	if x.httpWrapper == nil {
		x.httpWrapper = &util.HttpWrapper{}
	}
}

func (x *M3u8Controller) handleQueryQ(q string) string {
	if !util.IsHttpUrl(q) {
		buf, _ := base64.StdEncoding.DecodeString(q)
		q = string(buf)
	}
	return q
}

func (x *M3u8Controller) Proxy(ctx *gin.Context) {
	var q = x.handleQueryQ(ctx.Query("q"))

	if !util.IsHttpUrl(q) {
		ctx.JSON(http.StatusOK, gin.H{"code": "500", "msg": "请求格式错误"})
		return
	}

	log.Println("[Query]", q)

	resp, err := http.Head(q)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}
	//defer func() { _ = resp.Body.Close() }()

	var qContentType = resp.Header.Get(headers.ContentType)
	log.Println("[qContentType]", qContentType)

	switch qContentType {
	case "application/vnd.apple.mpegurl":
		fallthrough
	case "audio/x-mpegurl":
		fallthrough
	case "video/vnd.mpegurl":
		buf, _ := os.ReadFile("D:\\repo\\github.com\\lixiang4u\\airplayTV\\_debug\\1.m3u8")
		playlist, err := x.handleM3u8Url(buf)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "[error] %s", err.Error())
			return
		}
		ctx.Header(headers.ContentType, "application/vnd.apple.mpegurl")
		ctx.Header(headers.ContentDisposition, "inline; filename=playlist.m3u8")
		_, _ = ctx.Writer.WriteString(playlist.String())
		break
	case "application/octet-stream":
		req, err := http.NewRequest("GET", q, nil)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "[error] %s", err.Error())
			return
		}
		resp2, err := http.DefaultClient.Do(req)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "[error] %s", err.Error())
			return
		}
		defer func() { _ = resp2.Body.Close() }()

		ctx.Header(headers.ContentType, qContentType)
		_, _ = io.Copy(ctx.Writer, resp2.Body)
		break
	}
}

func (x *M3u8Controller) handleM3u8Url(m3u8Buff []byte) (m3u8.Playlist, error) {
	var proxyStreamUrl = "http://127.0.0.1:8099/api/m3u8p?q=%s"

	playList, listType, err := m3u8.DecodeFrom(bytes.NewBuffer(m3u8Buff), true)
	if err != nil {
		return playList, err
	}

	switch listType {
	case m3u8.MEDIA:
		mediapl := playList.(*m3u8.MediaPlaylist)
		for idx, val := range mediapl.Segments {
			if val == nil {
				continue
			}
			mediapl.Segments[idx].URI = fmt.Sprintf(proxyStreamUrl, base64.StdEncoding.EncodeToString([]byte(val.URI)))
		}
	case m3u8.MASTER:
		masterpl := playList.(*m3u8.MasterPlaylist)
		fmt.Printf("[BBBBB] %+v\n", masterpl)
	}

	return playList, nil
}
