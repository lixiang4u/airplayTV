package service

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/dengsgo/math-engine/engine"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"github.com/tidwall/gjson"
	"github.com/zc310/headers"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	czHost      = "https://czzy.top"
	czTagUrl    = "https://czzy.top/%s/movie_bt_series/dyy/page/%d"
	czSearchUrl = "https://czzy.top/daoyongjiekoshibushiyoubing?q=%s&f=_all&p=%d"
	czDetailUrl = "https://czzy.top/movie/%s.html"
	czPlayUrl   = "https://czzy.top/v_play/%s.html"
	ua          = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"
	m3u8pUrl    = "http://106.15.56.178:38386/api/m3u8p?q=%s"
)

//========================================================================
//==============================接口实现===================================
//========================================================================

type CZMovie struct {
	movie       Movie
	httpWrapper *util.HttpWrapper
	btVerifyUrl string
}

func (x *CZMovie) Init(movie Movie) {
	x.movie = movie
	if x.httpWrapper == nil {
		x.httpWrapper = &util.HttpWrapper{}
	}
	x.httpWrapper.SetHeader(headers.Origin, czHost)
	x.httpWrapper.SetHeader("authority", util.HandleHostname(czHost))
	x.httpWrapper.SetHeader(headers.Referer, czHost)
	x.httpWrapper.SetHeader(headers.UserAgent, ua)
	x.httpWrapper.SetHeader(headers.AcceptEncoding, "br, deflate, gzip")
}

func (x *CZMovie) ListByTag(tagName, page string) model.Pager {
	return x.czListByTag(tagName, page)
}

func (x *CZMovie) Search(search, page string) model.Pager {
	return x.czListBySearch(search, page)
}

func (x *CZMovie) Detail(id string) model.MovieInfo {
	return x.czVideoDetail(id)
}

func (x *CZMovie) Source(sid, vid string) model.Video {
	return x.czVideoSource(sid, vid)
}

//========================================================================
//==============================实际业务处理逻辑============================
//========================================================================

func (x *CZMovie) czListByTag(tagName, page string) model.Pager {
	_page, _ := strconv.Atoi(page)

	var pager = model.Pager{}
	pager.Limit = 25

	//err := x.btWaf()
	//if err != nil {
	//	log.Println("[绕过人机失败]", err.Error())
	//	return pager
	//}
	//b, err := x.httpWrapper.Get(fmt.Sprintf(czTagUrl, tagName, _page))
	//if err != nil {
	//	log.Println("[内容获取失败]", err.Error())
	//	return pager
	//}
	b, err := x.handleHttpRequestByM3u8p(fmt.Sprintf(czTagUrl, tagName, _page))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return pager
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return pager
	}
	doc.Find(".mi_cont .mi_ne_kd ul li").Each(func(i int, selection *goquery.Selection) {
		name := selection.Find(".dytit a").Text()
		tmpUrl, _ := selection.Find(".dytit a").Attr("href")
		thumb, _ := selection.Find("img.thumb").Attr("data-original")
		tag := selection.Find(".nostag").Text()
		actors := selection.Find(".inzhuy").Text()
		resolution := selection.Find(".hdinfo span").Text()

		pager.List = append(pager.List, model.MovieInfo{
			Id:         util.CZHandleUrlToId2(tmpUrl),
			Name:       name,
			Thumb:      thumb,
			Url:        tmpUrl,
			Actors:     strings.TrimSpace(actors),
			Tag:        tag,
			Resolution: resolution,
		})
	})

	doc.Find(".pagenavi_txt a").Each(func(i int, selection *goquery.Selection) {
		tmpHref, _ := selection.Attr("href")
		tmpList := strings.Split(tmpHref, "/")
		n, _ := strconv.Atoi(tmpList[len(tmpList)-1])
		if n*pager.Limit > pager.Total {
			pager.Total = n * pager.Limit
		}
	})

	pager.Current, _ = strconv.Atoi(doc.Find(".pagenavi_txt .current").Text())

	return pager
}

func (x *CZMovie) czListBySearch(query, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 20
	//err := x.btWaf()
	//if err != nil {
	//	log.Println("[绕过人机失败]", err.Error())
	//	return pager
	//}
	//x.httpWrapper.SetHeader(headers.Cookie, "esc_search_captcha=1; result=666;")
	//x.httpWrapper.SetHeader(headers.ContentType, "application/x-www-form-urlencoded")
	//h, b, err := x.httpWrapper.PostResponse(fmt.Sprintf(czSearchUrl, query, util.HandlePageNumber(page)), "result=666")
	//if err != nil {
	//	log.Println("[内容获取失败]", err.Error())
	//	return pager
	//}
	//b = x.btWafSearch(h, b, fmt.Sprintf(czSearchUrl, query, util.HandlePageNumber(page)))

	b, err := x.handleHttpRequestByM3u8p(fmt.Sprintf(czSearchUrl, query, util.HandlePageNumber(page)))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return pager
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return pager
	}

	doc.Find(".search_list ul li").Each(func(i int, selection *goquery.Selection) {
		name := selection.Find(".dytit a").Text()
		tmpUrl, _ := selection.Find(".dytit a").Attr("href")
		thumb, _ := selection.Find("img.thumb").Attr("data-original")
		tag := selection.Find(".nostag").Text()
		actors := selection.Find(".inzhuy").Text()

		pager.List = append(pager.List, model.MovieInfo{
			Id:     util.CZHandleUrlToId2(tmpUrl),
			Name:   name,
			Thumb:  thumb,
			Url:    tmpUrl,
			Actors: strings.TrimSpace(actors),
			Tag:    tag,
		})
	})

	doc.Find(".dytop .dy_tit_big span").Each(func(i int, selection *goquery.Selection) {
		if i == 0 {
			pager.Total, _ = strconv.Atoi(selection.Text())
		}
	})

	pager.Current, _ = strconv.Atoi(doc.Find(".pagenavi_txt .current").Text())

	return pager
}

func (x *CZMovie) czVideoDetail(id string) model.MovieInfo {
	var info = model.MovieInfo{}

	//err := x.btWaf()
	//if err != nil {
	//	log.Println("[绕过人机失败]", err.Error())
	//	return info
	//}
	//b, err := x.httpWrapper.Get(fmt.Sprintf(czDetailUrl, id))
	//if err != nil {
	//	log.Println("[内容获取失败]", err.Error())
	//	return info
	//}
	b, err := x.handleHttpRequestByM3u8p(fmt.Sprintf(czDetailUrl, id))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return info
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return info
	}

	doc.Find(".paly_list_btn a").Each(func(i int, selection *goquery.Selection) {
		tmpHref, _ := selection.Attr("href")
		info.Links = append(info.Links, model.Link{
			Id:    util.CZHandleUrlToId2(tmpHref),
			Name:  strings.ReplaceAll(selection.Text(), "厂长", ""),
			Url:   tmpHref,
			Group: "资源1",
		})
	})

	info.Id = id
	info.Thumb, _ = doc.Find(".dyxingq .dyimg img").Attr("src")
	info.Name = doc.Find(".dyxingq .moviedteail_tt h1").Text()
	info.Intro = strings.TrimSpace(doc.Find(".yp_context").Text())

	return info
}

func (x *CZMovie) czVideoSource(sid, vid string) model.Video {
	var video = model.Video{Id: sid}

	//err := x.btWaf()
	//if err != nil {
	//	log.Println("[绕过人机失败]", err.Error())
	//	return video
	//}
	//b, err := x.httpWrapper.Get(fmt.Sprintf(czPlayUrl, sid))
	//if err != nil {
	//	log.Println("[内容获取失败]", err.Error())
	//	return video
	//}

	b, err := x.handleHttpRequestByM3u8p(fmt.Sprintf(czPlayUrl, sid))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return video
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return video
	}

	var findLine = ""
	tmpList := strings.Split(string(b), "\n")
	for _, line := range tmpList {
		if strings.Contains(line, "md5.AES.decrypt") {
			findLine = line
			break
		}
	}
	if findLine != "" {
		video, err = x.czParseVideoSource(sid, findLine)

		bs, _ := json.MarshalIndent(video, "", "\t")
		log.Println(fmt.Sprintf("[video] %s", string(bs)))
		if err != nil {
			log.Println("[parse.video.error]", err)
		}
	}

	// 解析另一种iframe嵌套的视频
	iframeUrl, _ := doc.Find(".videoplay iframe").Attr("src")
	if strings.TrimSpace(iframeUrl) != "" {

		log.Println("[iframeUrl]", iframeUrl)

		if _, ok := util.RefererConfig[util.HandleHost(iframeUrl)]; ok {
			//需要chromedp加载后拿播放信息（数据通过js加密了）
			video.Source = iframeUrl
			video.Url = handleIframeEncrypedSourceUrl(iframeUrl)
		} else {
			// 直接可以拿到播放信息，或者需要解析加密的js数据得到信息
			video.Source, video.Type = x.getFrameUrlContents(iframeUrl)
			video.Source = util.FillUrlHost(video.Source, util.HandleHost(iframeUrl))
			if video.Source != "" {
				video.Url = video.Source
				//if util.CheckVideoUrl(video.Source) {
				//	video.Url = video.Source
				//} else {
				//	video.Url = HandleSrcM3U8FileToLocal(video.Id, video.Source, x.movie.IsCache)
				//}
			} else {
				video.Source = x.parseNetworkMediaUrl(fmt.Sprintf(czPlayUrl, sid))
				if util.CheckVideoUrl(video.Source) {
					video.Url = video.Source
				} else {
					video.Url = video.Source + "#default_parser"
				}
			}
			// 1、转为本地m3u8
			// 2、修改m3u8文件内容地址,支持跨域
		}
	}

	video.Name = doc.Find(".pclist .jujiinfo h3").Text()

	// 视频类型问题处理
	video = handleVideoType(video)

	return video
}

func (x *CZMovie) czParseVideoSource(id, js string) (model.Video, error) {
	var video = model.Video{}
	tmpList := strings.Split(strings.TrimSpace(js), ";")

	var data = ""
	var key = ""
	var iv = ""
	for index, str := range tmpList {
		if index == 0 {
			regex := regexp.MustCompile(`"\S+"`)
			data = strings.Trim(regex.FindString(str), `"`)
			continue
		}
		if index == 1 {
			regex := regexp.MustCompile(`"(\S+)"`)
			matchList := regex.FindStringSubmatch(str)
			if len(matchList) > 0 {
				key = matchList[len(matchList)-1]
			}
			continue
		}
		if index == 2 {
			regex := regexp.MustCompile(`\((\S+)\)`)
			matchList := regex.FindStringSubmatch(str)
			if len(matchList) > 0 {
				iv = matchList[len(matchList)-1]
			}
			continue
		}
	}

	log.Println(fmt.Sprintf("[parsing] key: %s, iv: %s", key, iv))

	if key == "" && data == "" {
		return video, errors.New("解析失败")
	}
	bs, err := util.DecryptByAes([]byte(key), []byte(iv), data)
	if err != nil {
		return video, errors.New("解密失败")
	}
	//tmpList = strings.Split(string(bs), "window")
	//if len(tmpList) < 1 {
	//	return video, errors.New("解密数据错误")
	//}

	regex := regexp.MustCompile(`video: {url: "(\S+?)",`)
	matchList := regex.FindStringSubmatch(string(bs))
	if len(matchList) < 1 {
		return video, errors.New("解析视频信息失败")
	}

	video.Id = id

	for index, m := range matchList {
		switch index {
		case 1:
			video.Source = m
			video.Url = m
			break
		case 2:
			video.Type = m
			break
		case 4:
			video.Thumb = m
			break
		default:
			break
		}
	}

	//  判断源文件是否需要下载
	if util.CheckVideoUrl(video.Source) {
		video.Url = video.Source
	} else {
		// 可能是m3u8，不返回hls就异常了，返回auto也不行啊
		if len(video.Type) == 0 {
			video.Type = "hls"
		}
		//video.Url = HandleSrcM3U8FileToLocal(id, video.Source, x.movie.IsCache)
		video.Url = video.Source
	}

	return video, nil
}

func handleVideoType(v model.Video) model.Video {
	// "https://yun.m3.c-zzy.online:1011/d/%E9%98%BF%E9%87%8C1%E5%8F%B7/%E8%8B%8F%E9%87%8C%E5%8D%97/Narco-Saints.S01E01.mp4?type=m3u8"

	tmpUrl, err := url.Parse(v.Source)
	if err != nil {
		return v
	}
	if util.StringInList(tmpUrl.Host, []string{"yun.m3.c-zzy.online:1011"}) {
		v.Type = "hls"
	}
	return v
}

func (x *CZMovie) getFrameUrlContents(frameUrl string) (sourceUrl, videoType string) {
	//sourceUrl = frameUrl
	videoType = "auto"

	// 请求这个域名需要（https://vavyuncz.cz01.org:83）
	x.httpWrapper.SetHeader("sec-fetch-mode", "navigate")
	x.httpWrapper.SetHeader("sec-fetch-dest", "iframe")

	bs, err := x.httpWrapper.Get(frameUrl)
	if err != nil {
		log.Println("[getFrameUrlContents.get.error]", err.Error())
		return
	}
	var htmlContents = string(bs)

	//log.Println("[=============htmlContents]", htmlContents)

	if strings.Contains(htmlContents, "var result_v2") && strings.Contains(htmlContents, "<script src=\"/js/player/index.min.js") {
		regEx1 := regexp.MustCompile(`var result_v2 = {"data":"(\S+?)"`)
		r1 := regEx1.FindStringSubmatch(htmlContents)
		if len(r1) > 1 {
			sourceUrl = x.parseEncryptedJsToUrl(r1[1])
		} else {
			log.Println("[iframe播放信息解析失败1]")
		}
	} else if strings.Contains(htmlContents, "var player") && strings.Contains(htmlContents, "var rand") {
		// 是否是AES加密数据
		regEx1 := regexp.MustCompile(`var rand = "(\S+)";`)
		regEx2 := regexp.MustCompile(`var player = "(\S+)";`)
		r1 := regEx1.FindStringSubmatch(htmlContents)
		r2 := regEx2.FindStringSubmatch(htmlContents)
		if len(r1) > 1 && len(r2) > 1 {
			buf, err := util.DecryptByAes([]byte("VFBTzdujpR9FWBhe"), []byte(r1[1]), r2[1])
			if err != nil {
				log.Println("[iframe播放信息解析失败2]", err.Error())
			} else {
				var result = gjson.ParseBytes(buf)
				sourceUrl = result.Get("url").String()
				videoType = util.GuessVideoType(sourceUrl)
			}
		} else {
			log.Println("[iframe播放信息解析失败2-2]", err.Error())
		}
	} else if strings.Contains(htmlContents, "sources:") {
		// 匹配播放文件
		regEx := regexp.MustCompile(`sources: \[{(\s+)src: '(\S+)',(\s+)type: '(\S+)'`)
		r := regEx.FindStringSubmatch(htmlContents)
		if len(r) >= 4 {
			sourceUrl = r[2]

			switch r[4] {
			case "application/vnd.apple.mpegurl":
				videoType = "hls"
			}
		} else {
			log.Println("[iframe播放信息解析失败3]")

		}
	}

	return
}

func handleIframeEncrypedSourceUrl(iframeUrl string) string {
	log.Println("[load.encrypted.iframe.video]")
	var err error

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var videoUrl string
	var videoUrlOk bool
	err = chromedp.Run(
		ctx,
		//chromedp.Navigate(iframeUrl),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, _, _, err := page.Navigate(iframeUrl).WithReferrer("https://www.czspp.com/").Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
		//chromedp.Evaluate(`urls;`, &res),
		// wait for footer element is visible (ie, page is loaded)
		// find and click "Example" link
		//chromedp.Click(`#example-After`, chromedp.NodeVisible),
		// retrieve the text of the textarea
		//chromedp.Value(`#div_player source`, &example),

		chromedp.WaitVisible(`#div_player`),

		chromedp.AttributeValue(`#div_player video source`, "src", &videoUrl, &videoUrlOk),
	)
	if err != nil && !strings.Contains(err.Error(), "net::ERR_ABORTED") {
		// Note: Ignoring the net::ERR_ABORTED page error is essential here
		// since downloads will cause this error to be emitted, although the
		// download will still succeed.
		log.Println("[network.error]", err)
		return ""
	}

	return videoUrl
}

func isWaf(html string) []byte {
	regEx := regexp.MustCompile(`window.location.href ="(\S+)";`)
	f := regEx.FindStringSubmatch(html)
	if len(f) < 2 {
		return nil
	}

	log.Println("[=========>]", fmt.Sprintf("%s%s", util.HandleHost(czHost), f[1]))

	resp, err := http.Get(fmt.Sprintf("%s%s", util.HandleHost(czHost), f[1]))
	if err != nil {
		log.Println("[IsWaf.error]", err.Error())
		return nil
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("[IsWaf.resp.body]", err.Error())
		return nil
	}
	log.Println("[IsWaf.error]", string(b))
	return b
}

func (x *CZMovie) GetVerifyUrl() string {
	b, err := x.httpWrapper.Get(czHost)
	if err != nil {
		log.Println("[访问主站错误]", err.Error())
		return ""
	}
	regEx := regexp.MustCompile(`<script type="text/javascript" src="(\S+)"></script>`)
	matchResult := regEx.FindStringSubmatch(string(b))

	if len(matchResult) < 2 {
		return ""
	}
	b, err = x.httpWrapper.Get(fmt.Sprintf("%s%s", strings.TrimRight(czHost, "/"), matchResult[1]))
	if err != nil {
		log.Println("[访问认证JS错误]", err.Error())
		return ""
	}

	regEx = regexp.MustCompile(`var key="(\w+)",value="(\w+)";`)
	matchResult2 := regEx.FindStringSubmatch(string(b))
	if len(matchResult2) < 3 {
		log.Println("[匹配认证配置错误] response:", string(b))
		return ""
	}
	log.Println("[解析验证配置]", util.ToJSON(matchResult2, true))

	regEx = regexp.MustCompile(`c.get\(\"(\S+)\&key\=`)
	matchResult3 := regEx.FindStringSubmatch(string(b))
	if len(matchResult3) < 2 {
		log.Println("[匹配认证地址错误] response:", string(b))
		return ""
	}
	log.Println("[解析验证地址]", util.ToJSON(matchResult3, true))

	tmpUrl := fmt.Sprintf(
		"%s%s&key=%s&value=%s",
		strings.TrimRight(czHost, "/"),    //域名
		matchResult3[1],                   // 接口地址
		matchResult2[1],                   // key
		x.btVerifyEncode(matchResult2[2]), // value
	)

	return tmpUrl
}

func (x *CZMovie) btVerifyEncode(value string) string {
	var tmpString string
	for _, v := range value {
		tmpString += fmt.Sprintf("%v", v)
	}
	return util.StringMd5(tmpString)
}

func (x *CZMovie) SetCookie() error {
	if x.btVerifyUrl == "" {
		x.btVerifyUrl = x.GetVerifyUrl()
	}

	log.Println("[btVerifyUrl]", x.btVerifyUrl)

	if x.btVerifyUrl == "" {
		return errors.New("解析人机认证失败")
	}
	h, body, err := x.httpWrapper.GetResponse(x.btVerifyUrl)

	if err != nil {
		x.btVerifyUrl = "" // 请求有问题，重置认证URL
		return err
	}
	tmpV := strings.TrimSpace(string(body))
	if v, ok := h["Set-Cookie"]; ok && strings.Contains(strings.TrimSpace(v[0]), tmpV) {
		x.httpWrapper.SetHeader("cookie", v[0])
		return nil
	}

	x.btVerifyUrl = "" // 请求返回数据，先重置认证URL吧

	return errors.New("没有发现cookie")
}

// 2022-12-05 新增验证规则
func (x *CZMovie) btWaf() error {
	b, err := x.httpWrapper.Get(czHost)
	if err != nil {
		log.Println("[访问主站错误]", err.Error())
		return err
	}

	if strings.Contains(string(b), "challenge-error-text") && strings.Contains(string(b), "cdn-cgi/challenge-platform") {
		return errors.New("[cloudflare challenge]")
	}

	regEx := regexp.MustCompile(`<script> window.location.href ="(\S+)"; </script>`)
	matchResult := regEx.FindStringSubmatch(string(b))

	if len(matchResult) < 2 {
		log.Println("[没有找到验证跳转/可能不存在验证]")
		return nil
	}
	tmpUrl := fmt.Sprintf("%s%s", strings.TrimRight(czHost, "/"), matchResult[1])
	b, err = x.httpWrapper.Get(tmpUrl)
	if err != nil {
		log.Println("[访问验证URL错误]", tmpUrl)
		return err
	}

	return nil
}

// 人机验证，计算
func (x *CZMovie) btWafSearch(h map[string][]string, html []byte, requestUrl string) []byte {
	// 第一次POST计算结果后会返回cookie，携带result=xxx的值
	// 再次POST第一次计算结果表单，写到如下两个cookie
	//cookie: esc_search_captcha=1
	//cookie: result=88
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return html
	}
	log.Println("===========title=======", doc.Find("title").Text())
	if strings.TrimSpace(doc.Find("title").Text()) != "人机验证" {
		log.Println("==========跳过人机验证的BUG========")
		return html
	}
	var mathText = doc.Find("form").Text()

	log.Println("[人机验证数据]", mathText)

	var result = 0.0
	result, err = engine.ParseAndExec(mathText[:strings.LastIndex(mathText, "=")])
	if err != nil {
		log.Println("[人机验证解析失败]", mathText, err.Error())
	}

	v, ok := h["Set-Cookie"]
	if ok {
		for _, s := range v {
			if strings.Contains(s, "PHPSESSID") {
				x.httpWrapper.SetHeader(headers.Cookie, s)
			}
		}
	}

	// 这里直接利用了个cookie检验漏洞，不做二次检验了
	//x.httpWrapper.SetHeader(headers.Cookie, "esc_search_captcha=1; result=47;")
	x.httpWrapper.SetHeader(headers.ContentType, "application/x-www-form-urlencoded")

	// 还他妈需要发两次请求什么鬼
	_, _ = x.httpWrapper.Post(requestUrl, fmt.Sprintf("result=%d", int(result)))
	b1, _ := x.httpWrapper.Post(requestUrl, fmt.Sprintf("result=%d", int(result)))

	return b1
}

// 从视频播放地址分析网络请求并找到媒体请求
func (x *CZMovie) parseNetworkMediaUrl(requestUrl string) string {
	var findUrl string

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			//log.Println("[network.EventRequestWillBeSent]", ev.Type, ev.Request.URL)
			if util.StringInList(ev.Type.String(), []string{"Media"}) {
				findUrl = ev.Request.URL
				cancel()
			}
		case *network.EventWebSocketCreated:
			//log.Println("[network.EventWebSocketCreated]", ev.URL)
		case *network.EventWebSocketFrameError:
			//log.Println("[network.EventWebSocketFrameError]", ev.ErrorMessage)
		case *network.EventWebSocketFrameSent:
			//log.Println("[network.EventWebSocketFrameSent]", ev.Response.PayloadData)
		case *network.EventWebSocketFrameReceived:
			//log.Println("[network.EventWebSocketFrameReceived]", ev.Response.PayloadData)
		case *network.EventResponseReceived:
			//log.Println("[network.EventResponseReceived]", ev.Type, ev.Response.URL)
		}
	})

	err := chromedp.Run(ctx,
		chromedp.Tasks{
			network.Enable(),
			chromedp.Navigate(requestUrl),
			chromedp.WaitVisible("#I_FUCK_YOU"), // 等一个不存在的节点，然后通过event中cancel()接下来的所有request
		},
	)

	if err != nil {
		log.Println("[chromedp.Run.Error]", err.Error())
	}

	return findUrl
}

func (x *CZMovie) parseEncryptedJsToUrl(result_v2 string) string {
	// htoStr
	var chars = strings.Split(result_v2, "")
	slices.Reverse(chars)
	var sb = strings.Builder{}
	var tmpStr = ""
	var buf []byte
	var err error
	for i := 0; i < len(chars); i += 2 {
		tmpStr = chars[i] + chars[i+1]
		buf, err = hex.DecodeString(tmpStr)
		if err != nil {
			log.Println("[decodeHexError]", err.Error())
			break
		}
		sb.Write(buf)
	}

	// decodeStr
	var tmpUrl = sb.String()
	var tmpA = (len(tmpUrl) - 7) / 2

	return fmt.Sprintf("%s%s", tmpUrl[0:tmpA], tmpUrl[tmpA+7:])
}

func (x *CZMovie) X(requestUrl string) string {
	return x.fuckCfClearance(requestUrl)
}

func (x *CZMovie) fuckCfClearance(requestUrl string) string {
	var findUrl string

	allocCtx, allocCancel := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.UserAgent(ua),
	)
	defer allocCancel()

	ctx, ctxCancel := chromedp.NewContext(allocCtx)
	defer ctxCancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, timeoutCancel := context.WithTimeout(ctx, 20*time.Second)
	defer timeoutCancel()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			//log.Println("[network.EventRequestWillBeSent]", ev.Type, ev.Request.URL)
			//if util.StringInList(ev.Type.String(), []string{"Stylesheet", "Image", "Font"}) {
			//	ev.Request.URL = ""
			//}
		case *network.EventWebSocketCreated:
			//log.Println("[network.EventWebSocketCreated]", ev.URL)
		case *network.EventWebSocketFrameError:
			log.Println("[network.EventWebSocketFrameError]", ev.ErrorMessage)
		case *network.EventWebSocketFrameSent:
			//log.Println("[network.EventWebSocketFrameSent]", ev.Response.PayloadData)
		case *network.EventWebSocketFrameReceived:
			//log.Println("[network.EventWebSocketFrameReceived]", ev.Response.PayloadData)
		case *network.EventResponseReceived:
			//log.Println("[network.EventResponseReceived]", ev.Type, ev.Response.URL)
			//if ev.Type == network.ResourceTypeDocument {
			//	log.Println("[ev.Response.Headers]", ev.Response.URL, util.ToJSON(ev.Response.Headers, true))
			//	network.GetResponseBody(ev.RequestID)
			//}
			//log.Println("[ev.Response.Headers]", ev.Response.URL, util.ToJSON(ev.Response.Headers, true))
			//log.Println("[===============>Header]", util.ToJSON(ev.Response.Headers, true))
		case *runtime.EventConsoleAPICalled:
			//log.Println("[runtime.EventConsoleAPICalled]", ev.Type, util.ToJSON(ev.Args, true))
			//for _, arg := range ev.Args {
			//	fmt.Printf("[EventConsoleAPICalled] %s - %s\n", arg.Type, arg.Value)
			//}

		}
	})

	//var res []byte
	//var html string
	//var iframes []*cdp.Node
	//log.Println("=========0")
	//err := chromedp.Run(
	//	ctx,
	//	network.Enable(),
	//	chromedp.Navigate(requestUrl),
	//)
	//if err != nil {
	//	log.Println("[Error1]", err.Error())
	//	return ""
	//}
	//log.Println("=========1")
	//err = chromedp.Run(
	//	ctx,
	//	chromedp.WaitReady("iframe"),
	//	chromedp.Nodes("iframe", &iframes, chromedp.ByQuery),
	//	chromedp.ActionFunc(func(ctx context.Context) error {
	//
	//		log.Println("[iframes]", len(iframes))
	//
	//		return nil
	//	}),
	//)
	//if err != nil {
	//	log.Println("[Error2]", err.Error())
	//	return ""
	//}
	//log.Println("=========2")
	//err = chromedp.Run(
	//	ctx,
	//	chromedp.ActionFunc(func(ctx context.Context) error {
	//		log.Println("[ActionFunc] 1")
	//		log.Println("[ActionFunc]", util.ToJSON(iframes[0], true))
	//		return nil
	//	}),
	//	chromedp.WaitReady(".main-wrapper", chromedp.ByQuery, chromedp.FromNode(iframes[0])),
	//	//chromedp.WaitVisible("body", chromedp.ByQuery, chromedp.FromNode(iframes[0])),
	//	//chromedp.InnerHTML(".ctp-label", &html),
	//	chromedp.ActionFunc(func(ctx context.Context) error {
	//		log.Println("[ActionFunc] 2")
	//		return nil
	//	}),
	//
	//	chromedp.InnerHTML(".main-wrapper", &html, chromedp.ByQuery),
	//	chromedp.FullScreenshot(&res, 90),
	//)

	var isWebDriver bool
	var screenshot []byte
	var cookie string
	err := chromedp.Run(
		ctx,
		// click 56，290
		chromedp.EmulateViewport(880, 435),
		chromedp.Tasks{
			network.Enable(),
			chromedp.Navigate(requestUrl),
			chromedp.Sleep(time.Second * 3),
			chromedp.Evaluate(`window.navigator.webdriver`, &isWebDriver),
			chromedp.MouseClickXY(56, 290),
			chromedp.MouseClickXY(60, 290),
			chromedp.WaitVisible(".mikd"),
			chromedp.ActionFunc(func(ctx context.Context) error {
				cookies, _ := network.GetAllCookies().Do(ctx)
				cookie = x.parseCookie(cookies)
				return nil
			}),
			chromedp.FullScreenshot(&screenshot, 90),
		},
	)

	log.Println("[isWebDriver]", isWebDriver)
	log.Println("[cookie]", cookie)

	if err != nil {
		log.Println("[chromedp.Run.Error]", err.Error())
	}
	if err := os.WriteFile(filepath.Join(util.AppPath(), fmt.Sprintf("fullScreenshot-%d.png", time.Now().Unix())), screenshot, fs.ModePerm); err != nil {
		log.Fatal(err)
	}

	return findUrl
}

func (x *CZMovie) parseCookie(cookies []*network.Cookie) string {
	var cookieString = ""
	if cookies == nil || len(cookies) <= 0 {
		return cookieString
	}
	for _, cookie := range cookies {
		if len(cookieString) <= 0 {
			cookieString = fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
		} else {
			cookieString = fmt.Sprintf("%s; %s=%s", cookieString, cookie.Name, cookie.Value)
		}
	}
	return cookieString
}

func (x *CZMovie) handleHttpRequestByM3u8p(requestUrl string) ([]byte, error) {
	header, buf, err := x.httpWrapper.GetResponse(fmt.Sprintf(m3u8pUrl, requestUrl))
	if err != nil {
		return nil, err
	}
	log.Println("[ContentType]", header.Get(headers.ContentType), requestUrl)
	return buf, nil
}
