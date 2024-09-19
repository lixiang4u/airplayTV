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
	"net/http"
	"strings"
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
func (x *M3u8Controller) getRequestUrlF(ctx *gin.Context) string {
	if ctx.Request.TLS != nil {
		return fmt.Sprintf("https://%s/%s?q=%%s", ctx.Request.Host, strings.TrimLeft(ctx.Request.URL.Path, "/"))
	} else {
		return fmt.Sprintf("http://%s/%s?q=%%s", ctx.Request.Host, strings.TrimLeft(ctx.Request.URL.Path, "/"))
	}
}

func (x *M3u8Controller) Proxy(ctx *gin.Context) {
	var q = x.handleQueryQ(ctx.Query("q"))

	if !util.IsHttpUrl(q) {
		ctx.String(http.StatusInternalServerError, "[error] 请求格式错误")
		return
	}

	//log.Println("[Query]", q)

	resp, err := http.Head(q)
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("[error] %s", err.Error()))
		return
	}
	//defer func() { _ = resp.Body.Close() }()

	var qContentType = resp.Header.Get(headers.ContentType)
	//log.Println("[qContentType]", qContentType)

	switch qContentType {
	case "application/vnd.apple.mpegurl":
		fallthrough
	case "audio/x-mpegurl":
		fallthrough
	case "video/vnd.mpegurl":
		_, buf, err := x.httpWrapper.GetResponse(q)
		if err != nil {
			ctx.String(http.StatusInternalServerError, fmt.Sprintf("[error] %s", err.Error()))
			return
		}

		playlist, err := x.handleM3u8Url(ctx, buf)
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

func (x *M3u8Controller) handleM3u8Url(ctx *gin.Context, m3u8Buff []byte) (m3u8.Playlist, error) {
	var proxyStreamUrl = x.getRequestUrlF(ctx)

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
