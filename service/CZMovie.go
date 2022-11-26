package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	czHost      = "https://www.czspp.com"
	czTagUrl    = "https://www.czspp.com/%s/page/%d"
	czSearchUrl = "https://www.czspp.com/xssearch?q=%s&p=%d"
	czDetailUrl = "https://www.czspp.com/movie/%s.html"
	czPlayUrl   = "https://www.czspp.com/v_play/%s.html"
)

//========================================================================
//==============================接口实现===================================
//========================================================================

type CZMovie struct {
	movie       Movie
	httpWrapper *util.HttpWrapper
}

func (x *CZMovie) Init(movie Movie) {
	x.movie = movie
	if x.httpWrapper == nil {
		x.httpWrapper = &util.HttpWrapper{}
	}
	x.httpWrapper.SetHeader("origin", czHost)
	x.httpWrapper.SetHeader("authority", util.HandleHostname(czHost))
	x.httpWrapper.SetHeader("referer", czHost)
	x.httpWrapper.SetHeader("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36")
	x.httpWrapper.SetHeader("cookie", "")
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

	err := x.SetCookie()
	if err != nil {
		log.Println("[绕过人机失败]", err.Error())
		return pager
	}
	b, err := x.httpWrapper.Get(fmt.Sprintf(czTagUrl, tagName, _page))
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
			Id:         util.CZHandleUrlToId(tmpUrl),
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

	err := x.SetCookie()
	if err != nil {
		log.Println("[绕过人机失败]", err.Error())
		return pager
	}
	b, err := x.httpWrapper.Get(fmt.Sprintf(czSearchUrl, query, util.HandlePageNumber(page)))
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
			Id:     util.CZHandleUrlToId(tmpUrl),
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

	_ = x.SetCookie()
	c := x.movie.NewColly()

	c.OnHTML(".paly_list_btn", func(element *colly.HTMLElement) {
		element.ForEach("a", func(i int, element *colly.HTMLElement) {
			info.Links = append(info.Links, model.Link{
				Id:   util.CZHandleUrlToId2(element.Attr("href")),
				Name: strings.ReplaceAll(element.Text, "厂长", ""),
				Url:  element.Attr("href"),
			})
		})
	})

	c.OnHTML(".dyxingq", func(element *colly.HTMLElement) {
		info.Thumb = element.ChildAttr(".dyimg img", "src")
		info.Name = element.ChildText(".moviedteail_tt h1")
	})

	c.OnHTML(".yp_context", func(element *colly.HTMLElement) {
		info.Intro = strings.TrimSpace(element.Text)
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})
	c.OnResponse(func(response *colly.Response) {
		if newResp := isWaf(string(response.Body)); newResp != nil {
			response.Body = newResp
		}
	})

	err := c.Visit(fmt.Sprintf(czDetailUrl, id))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return info
}

func (x *CZMovie) czVideoSource(sid, vid string) model.Video {
	var video = model.Video{Id: sid}
	var err error

	_ = x.SetCookie()
	c := x.movie.NewColly()

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	c.OnResponse(func(response *colly.Response) {
		if newResp := isWaf(string(response.Body)); newResp != nil {
			response.Body = newResp
		}

		var findLine = ""
		tmpList := strings.Split(string(response.Body), "\n")
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
	})

	// 解析另一种iframe嵌套的视频
	c.OnHTML(".videoplay iframe", func(element *colly.HTMLElement) {
		iframeUrl := element.Attr("src")
		log.Println("======[iframeUrl] ", iframeUrl)

		if _, ok := util.RefererConfig[util.HandleHost(iframeUrl)]; ok {
			//需要chromedp加载后拿播放信息（数据通过js加密了）
			video.Source = iframeUrl
			video.Url = handleIframeEncrypedSourceUrl(iframeUrl)
		} else {
			// 直接可以拿到播放信息
			video.Source, video.Type = getFrameUrlContents(iframeUrl)
			video.Url = HandleSrcM3U8FileToLocal(video.Id, video.Source, x.movie.IsCache)
			// 1、转为本地m3u8
			// 2、修改m3u8文件内容地址,支持跨域
		}
	})

	c.OnHTML(".jujiinfo", func(element *colly.HTMLElement) {
		video.Name = element.ChildText("h3")
	})
	c.OnResponse(func(response *colly.Response) {
		if newResp := isWaf(string(response.Body)); newResp != nil {
			response.Body = newResp
		}
	})

	err = c.Visit(fmt.Sprintf(czPlayUrl, sid))
	if err != nil {
		log.Println("[ERR]", err.Error())
	}

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
	tmpList = strings.Split(string(bs), "window")
	if len(tmpList) < 1 {
		return video, errors.New("解密数据错误")
	}

	regex := regexp.MustCompile(`{url: "(\S+)",type:"(\S+)",([\S\s]*)pic:'(\S+)'}`)
	matchList := regex.FindStringSubmatch(tmpList[0])

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

	video.Url = HandleSrcM3U8FileToLocal(id, video.Source, x.movie.IsCache)

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

func getFrameUrlContents(frameUrl string) (sourceUrl, videoType string) {
	sourceUrl = frameUrl
	videoType = "auto"

	resp, err := http.Get(frameUrl)
	if err != nil {
		log.Println("[getFrameUrlContents.get.error]", err.Error())
		return
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("[getFrameUrlContents.body.error]", err.Error())
		return
	}

	// 匹配播放文件
	regEx := regexp.MustCompile(`sources: \[{(\s+)src: '(\S+)',(\s+)type: '(\S+)'`)
	r := regEx.FindStringSubmatch(string(bs))
	if len(r) < 4 {
		return
	}
	sourceUrl = r[2]

	switch r[4] {
	case "application/vnd.apple.mpegurl":
		videoType = "hls"
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

func (x *CZMovie) SetCookie() error {
	tmpUrl := "https://www.czspp.com/a20be899_96a6_40b2_88ba_32f1f75f1552_yanzheng_ip.php?type=96c4e20a0e951f471d32dae103e83881&key=21ce3b4f9c0d19a7797e28e44824be3b&value=11098d503592721f9914770753951607"
	h, body, err := x.httpWrapper.GetResponse(tmpUrl)

	if err != nil {
		return err
	}
	tmpV := strings.TrimSpace(string(body))
	if v, ok := h["Set-Cookie"]; ok && strings.Contains(strings.TrimSpace(v[0]), tmpV) {
		x.httpWrapper.SetHeader("cookie", v[0])
		return nil
	}

	return errors.New("没有发现cookie")
}
