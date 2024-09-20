package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/grafov/m3u8"
	"github.com/lixiang4u/airplayTV/service"
	"github.com/lixiang4u/airplayTV/util"
	"github.com/zc310/headers"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
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
		q = x.base64DecodingX(q)
	}
	return q
}

func (x *M3u8Controller) getRequestUrlF(ctx *gin.Context) string {
	return fmt.Sprintf("https://%s/%s?q=%%s", ctx.Request.Host, strings.TrimLeft(ctx.Request.URL.Path, "/"))
}

func (x *M3u8Controller) Proxy(ctx *gin.Context) {
	var q = x.handleQueryQ(ctx.Query("q"))

	if !util.IsHttpUrl(q) {
		ctx.String(http.StatusInternalServerError, "参数错误")
		return
	}

	log.Println("[Query]", q)

	resp, err := http.Head(q)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}

	var qContentType = strings.ToLower(resp.Header.Get(headers.ContentType))

	if resp.StatusCode == 200 {
		x.handleResponseContentType(ctx, q, qContentType)
		return
	} else {
		// 可能是禁止Head，需要GET一次
		x.handleM3u8Stream(ctx, q, qContentType)
	}
}

func (x *M3u8Controller) handleM3u8Url(ctx *gin.Context, m3u8Url string, m3u8Buff []byte) (m3u8.Playlist, error) {
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
			val.URI = x.handleM3u8PlayListUrl(val.URI, m3u8Url)
			mediapl.Segments[idx].URI = fmt.Sprintf(proxyStreamUrl, x.base64EncodingX(val.URI))
		}
	case m3u8.MASTER:
		masterpl := playList.(*m3u8.MasterPlaylist)
		for idx, val := range masterpl.Variants {
			if val == nil {
				continue
			}
			val.URI = x.handleM3u8PlayListUrl(val.URI, m3u8Url)
			masterpl.Variants[idx].URI = fmt.Sprintf(proxyStreamUrl, x.base64EncodingX(val.URI))
		}
	}

	return playList, nil
}

func (x *M3u8Controller) base64EncodingX(q string) string {
	return base64.StdEncoding.EncodeToString(
		[]byte(util.ToJSON(
			map[string]interface{}{"q": q},
			false,
		)),
	)
}

func (x *M3u8Controller) base64DecodingX(q string) string {
	if util.IsHttpUrl(q) {
		return q
	}
	buf, err := base64.StdEncoding.DecodeString(q)
	if err != nil {
		return ""
	}
	var m map[string]interface{}
	if err = json.Unmarshal(buf, &m); err != nil {
		return ""
	}
	if v, ok := m["q"]; ok {
		return v.(string)
	}
	return ""
}

func (x *M3u8Controller) handleM3u8PlayListUrl(playUrl, m3u8Url string) string {
	if util.IsHttpUrl(playUrl) {
		return playUrl
	}
	playUrl = strings.TrimSpace(playUrl)
	if strings.HasPrefix(playUrl, "/") {
		parsedUrl, _ := url.Parse(m3u8Url)
		return fmt.Sprintf("%s://%s/%s", parsedUrl.Scheme, parsedUrl.Host, playUrl)
	} else {
		parsedUrl, _ := url.Parse(m3u8Url)
		return fmt.Sprintf("%s://%s/%s/%s", parsedUrl.Scheme, parsedUrl.Host, strings.TrimLeft(filepath.Dir(parsedUrl.Path), "\\/"), playUrl)
	}
}

func (x *M3u8Controller) handleResponseM3u8PlayList(ctx *gin.Context, q string, qResponse []byte) {
	if qResponse == nil {
		var err error
		_, qResponse, err = x.httpWrapper.GetResponse(q)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}
	}

	playlist, err := x.handleM3u8Url(ctx, q, qResponse)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	ctx.Header(headers.ContentType, "application/vnd.apple.mpegurl")
	ctx.Header(headers.ContentDisposition, fmt.Sprintf("inline; filename=playlist%d.m3u8", time.Now().Unix()))
	_, _ = ctx.Writer.WriteString(playlist.String())
}

func (x *M3u8Controller) handleM3u8Stream(ctx *gin.Context, q, qContentType string) {
	req, err := http.NewRequest("GET", q, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	defer func() { _ = resp2.Body.Close() }()
	switch strings.ToLower(resp2.Header.Get(headers.ContentType)) {
	case "application/vnd.apple.mpegurl":
		fallthrough
	case "application/apple.vnd.mpegurl":
		fallthrough
	case "application/x-mpegurl":
		fallthrough
	case "audio/x-mpegurl":
		fallthrough
	case "video/vnd.mpegurl":
		buf, err := io.ReadAll(resp2.Body)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}
		x.handleResponseM3u8PlayList(ctx, q, buf)
		return
	}

	ctx.Header(headers.ContentType, qContentType)
	_, _ = io.Copy(ctx.Writer, resp2.Body)
}

func (x *M3u8Controller) handleResponseContentType(ctx *gin.Context, q, qContentType string) {
	switch qContentType {
	case "application/vnd.apple.mpegurl":
		fallthrough
	case "application/apple.vnd.mpegurl":
		fallthrough
	case "application/x-mpegurl":
		fallthrough
	case "audio/x-mpegurl":
		fallthrough
	case "video/vnd.mpegurl":
		x.handleResponseM3u8PlayList(ctx, q, nil)
		break
	case "video/mp2t": // ts文件
		fallthrough
	case "image/jpeg": // 图像文件替代ts
		fallthrough
	case "application/octet-stream":
		x.handleM3u8Stream(ctx, q, qContentType)
		break
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": fmt.Sprintf("无法解析(%s)", qContentType),
		})
		break
	}

}
