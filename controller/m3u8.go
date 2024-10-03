package controller

import (
	"bufio"
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
	ctx.Header("X-Original-Url", q)

	resp, err := http.Head(q)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}

	if resp.StatusCode == 200 {
		x.handleResponseContentType(ctx, q, strings.ToLower(resp.Header.Get(headers.ContentType)))
		return
	} else {
		// 可能是禁止Head，需要GET一次
		x.handleM3u8Stream(ctx, q)
	}
}

func (x *M3u8Controller) handleM3u8Url(ctx *gin.Context, m3u8Url string, m3u8Buff []byte) (m3u8.Playlist, error) {
	var proxyStreamUrl = x.getRequestUrlF(ctx)

	playList, listType, err := m3u8.DecodeFrom(bytes.NewBuffer(m3u8Buff), true)
	if err != nil {
		return playList, err
	}
	var m3u8Host = util.HandleHostname(m3u8Url) // 不带端口的域名

	switch listType {
	case m3u8.MEDIA:
		mediapl := playList.(*m3u8.MediaPlaylist)
		if mediapl.Key != nil && len(mediapl.Key.URI) > 0 {
			mediapl.Key.URI = fmt.Sprintf(proxyStreamUrl, x.base64EncodingX(x.handleM3u8PlayListUrl(mediapl.Key.URI, m3u8Url)))
		}
		for idx, val := range mediapl.Segments {
			// 过滤广告URL
			val = x.handleMediaSegmentAdvertisement(val, m3u8Host)
			if val == nil {
				continue
			}
			// 修正URL路径带域名
			val.URI = x.handleM3u8PlayListUrl(val.URI, m3u8Url)
			// 设置代理URL，如果URL空（广告过滤会设置为空）则调过
			if len(val.URI) > 0 {
				mediapl.Segments[idx].URI = fmt.Sprintf(proxyStreamUrl, x.base64EncodingX(val.URI))
			}
			// 这里不设置nil会导致出现两个EXT-X-KEY字段
			val.Key = nil
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
	if len(playUrl) == 0 {
		return playUrl
	}
	if util.IsHttpUrl(playUrl) {
		return playUrl
	}
	playUrl = strings.TrimSpace(playUrl)
	if strings.HasPrefix(playUrl, "/") {
		parsedUrl, _ := url.Parse(m3u8Url)
		return fmt.Sprintf("%s://%s/%s", parsedUrl.Scheme, parsedUrl.Host, strings.TrimLeft(playUrl, "/"))
	} else {
		parsedUrl, _ := url.Parse(m3u8Url)
		return fmt.Sprintf(
			"%s://%s/%s/%s",
			parsedUrl.Scheme,
			parsedUrl.Host,
			strings.ReplaceAll(strings.TrimLeft(filepath.Dir(parsedUrl.Path), "\\/"), "\\", "/"),
			playUrl,
		)
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

func (x *M3u8Controller) handleM3u8Stream(ctx *gin.Context, q string) {
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

	var qContentType = strings.ToLower(resp2.Header.Get(headers.ContentType))

	//bytes.NewReader()
	// 小文件直接解析内容
	var respReader io.Reader
	if util.StringToInt(resp2.Header.Get(headers.ContentLength)) < 1024*800 {
		respBuff, err := io.ReadAll(resp2.Body)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}
		_, err = x.handleM3u8Url(ctx, q, respBuff)
		if err == nil {
			// 实际是播放文件，需要串改Content-Type
			qContentType = "application/x-mpegurl"
		}
		respReader = bytes.NewReader(respBuff)
	} else {
		respReader = bufio.NewReader(resp2.Body)
	}

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
		buf, _ := io.ReadAll(respReader)
		x.handleResponseM3u8PlayList(ctx, q, buf)
		return
	}

	ctx.Header(headers.ContentType, qContentType)
	_, _ = io.Copy(ctx.Writer, respReader)
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
	case "image/png": // 图像文件替代ts
		fallthrough
	case "": // https://p.hhwenjian.com:65/hls/488/20240824/2810990/plist0.ts
		fallthrough
	case "application/octet-stream":
		x.handleM3u8Stream(ctx, q)
		break
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": fmt.Sprintf("无法解析(%s)", qContentType),
		})
		break
	}

}

func (x *M3u8Controller) handleMediaSegmentAdvertisement(segment *m3u8.MediaSegment, m3u8Host string) *m3u8.MediaSegment {
	if segment == nil {
		return nil
	}
	switch m3u8Host {
	case "c1.rrcdnbf3.com":
		if strings.Contains(segment.URI, "video/adjump") {
			segment.Duration = 0
			segment.URI = ""
			return segment
		}
	case "debug.rrcdnbf3.com":
		if strings.Contains(segment.URI, "video/adjump") {
			segment.Duration = 0
			segment.URI = ""
			return segment
		}
	}

	return segment
}
