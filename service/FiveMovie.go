package service

import "C"
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/dop251/goja"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"github.com/tidwall/gjson"
	"github.com/zc310/headers"
	"log"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	fiveHost      = "https://555movie.me"
	fiveTagUrl    = "https://555movie.me/vodshow/1--------%d---.html"
	fiveSearchUrl = "https://555movie.me/vodsearch/%s----------%d---.html"
	fiveDetailUrl = "https://555movie.me/voddetail/%s.html"
	fivePlayUrl   = "https://555movie.me/vodplay/%s.html"
)

//========================================================================
//==============================接口实现===================================
//========================================================================

type FiveMovie struct {
	Movie
	httpWrapper *util.HttpWrapper
	btVerifyUrl string
}

func (x *FiveMovie) Init(movie Movie) {
	x.Movie = movie
	if x.httpWrapper == nil {
		x.httpWrapper = &util.HttpWrapper{}
	}
	x.httpWrapper.SetHeader(headers.Origin, fiveHost)
	x.httpWrapper.SetHeader("Host", util.HandleHostname(fiveHost))
	x.httpWrapper.SetHeader(headers.UserAgent, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36")
}

func (x *FiveMovie) ListByTag(tagName, page string) model.Pager {
	return x.fiveListByTag(tagName, page)
}

func (x *FiveMovie) Search(search, page string) model.Pager {
	return x.fiveListBySearch(search, page)
}

func (x *FiveMovie) Detail(id string) model.MovieInfo {
	return x.fiveVideoDetail(id)
}

func (x *FiveMovie) Source(sid, vid string) model.Video {
	return x.fiveVideoSource(sid, vid)
}

//========================================================================
//==============================实际业务处理逻辑============================
//========================================================================

func (x *FiveMovie) fiveListByTag(tagName, page string) model.Pager {
	_page := util.HandlePageNumber(page)

	var pager = model.Pager{}
	pager.Limit = 16

	//_, _ = x.JA3Request(fmt.Sprintf(fiveTagUrl, _page))

	// 还必须有这个多余动作，不然colly需要设置Host头
	_, err := x.httpWrapper.Get(fmt.Sprintf(fiveTagUrl, _page))
	if err != nil {
		log.Println("[fiveListByTag.error]", err.Error())
		return pager
	}

	c := x.Movie.NewColly()

	c.OnHTML(".module-items a.module-item", func(element *colly.HTMLElement) {

		name := element.ChildText(".module-poster-item-title")
		url := element.Attr("href")
		thumb := element.ChildAttr("img.lazy", "data-original")
		tag := element.ChildText(".module-item-note")
		actors := element.ChildText(".xxx")
		resolution := element.ChildText(".module-item-note")

		pager.List = append(pager.List, model.MovieInfo{
			Id:         util.CZHandleUrlToId(url),
			Name:       name,
			Thumb:      thumb,
			Url:        url,
			Actors:     actors,
			Tag:        tag,
			Resolution: resolution,
		})
	})

	c.OnHTML("#page", func(element *colly.HTMLElement) {
		element.ForEach("a.page-next", func(i int, element *colly.HTMLElement) {
			tmpList := strings.Split(element.Attr("href"), "/")
			n, _ := strconv.Atoi(strings.TrimRight(tmpList[len(tmpList)-1], ".html"))
			if n*pager.Limit > pager.Total {
				pager.Total = n * pager.Limit
			}
		})
	})

	c.OnHTML("#page .page-current", func(element *colly.HTMLElement) {
		pager.Current, _ = strconv.Atoi(element.Text)
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
		request.Headers.Set("Host", util.HandleHostname(fiveHost))
	})
	c.OnResponse(func(response *colly.Response) {
		if newResp := isWaf(string(response.Body)); newResp != nil {
			response.Body = newResp
		}
	})

	err = c.Visit(fmt.Sprintf(fiveTagUrl, _page))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return pager
}

func (x *FiveMovie) fiveListBySearch(query, page string) model.Pager {
	var requestUrl = fmt.Sprintf(fiveSearchUrl, query, util.HandlePageNumber(page))
	var pager = model.Pager{}
	pager.Limit = 16

	c := x.Movie.NewColly()

	c.OnHTML(".module-items .module-card-item", func(element *colly.HTMLElement) {
		name := element.ChildText(".module-card-item-title a strong")
		url := element.ChildAttr(".module-card-item-title a", "href")
		thumb := element.ChildAttr(".module-item-pic .lazyload", "data-original")
		tag := element.ChildText(".module-item-note")
		actors := element.ChildText(".module-info-item-content")

		pager.List = append(pager.List, model.MovieInfo{
			Id:     util.CZHandleUrlToId(url),
			Name:   name,
			Thumb:  thumb,
			Url:    url,
			Actors: actors,
			Tag:    tag,
		})
	})

	c.OnHTML("#page", func(element *colly.HTMLElement) {
		// /vodsearch/WE----------51---.html
		element.ForEach("a", func(i int, element *colly.HTMLElement) {

			tmpList := strings.Split(element.Attr("href"), "-")
			if len(tmpList) < 4 {
				return
			}
			pager.Total, _ = strconv.Atoi(tmpList[len(tmpList)-4])
		})
	})

	pager.Current = util.HandlePageNumber(page)

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})
	c.OnResponse(func(response *colly.Response) {
		log.Println("[responseBody]", string(response.Body))
		x.checkFiveWaf(requestUrl, response.Body)
	})

	err := c.Visit(requestUrl)
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return pager
}

func (x *FiveMovie) fiveVideoDetail(id string) model.MovieInfo {
	var info = model.MovieInfo{Id: id}
	var idxList = make([][]string, 0)

	c := x.Movie.NewColly()

	c.OnHTML(".module-tab-items-box", func(element *colly.HTMLElement) {
		element.ForEach(".module-tab-item", func(i0 int, element *colly.HTMLElement) {
			idxList = append(idxList, []string{strconv.Itoa(i0), element.ChildText("span")})
		})
	})

	c.OnHTML(".module-info-heading", func(element *colly.HTMLElement) {
		info.Name = element.ChildText("h1")
	})
	c.OnHTML(".module-info-poster .module-item-pic .lazyload", func(element *colly.HTMLElement) {
		info.Thumb = element.Attr("data-original")
	})
	c.OnHTML(".module-info-content .module-info-introduction-content p", func(element *colly.HTMLElement) {
		info.Intro = strings.TrimSpace(element.Text)
	})

	c.OnHTML("body", func(element *colly.HTMLElement) {
		element.ForEach(".module-list", func(i int, element *colly.HTMLElement) {
			for _, tmpValue := range idxList {
				if strconv.Itoa(i) != tmpValue[0] {
					continue
				}
				element.ForEach(".module-play-list-link", func(i int, element *colly.HTMLElement) {
					info.Links = append(info.Links, model.Link{
						Id:    util.CZHandleUrlToId2(element.Attr("href")),
						Name:  element.ChildText("span"),
						Url:   element.Attr("href"),
						Group: tmpValue[1],
					})
				})
			}
		})
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf(fiveDetailUrl, id))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return info
}

func (x *FiveMovie) fiveVideoSource(sid, vid string) model.Video {
	var video = model.Video{Id: sid}

	//video.Source = x.fiveParseVideoUrl(fmt.Sprintf(fivePlayUrl, sid))
	var err error
	video.Source, err = x.fiveDecryptVideoUrl(sid)
	if err != nil {
		log.Println("[fiveDecryptVideoUrlError]", sid, err.Error())
		return video
	}

	video.Url = HandleSrcM3U8FileToLocal(sid, video.Source, x.Movie.IsCache)

	// 视频类型问题处理
	video = x.handleVideoType(video)

	return video
}

func (x *FiveMovie) fiveDecryptVideoUrl(sid string) (string, error) {
	buf, err := x.httpWrapper.Get(fmt.Sprintf(fivePlayUrl, sid))
	if err != nil {
		return "", err
	}
	var findStr = util.SimpleRegEx(string(buf), `"url":"(\S+?)","url_next"`)

	// 文件：https://3d-platform-pro.obs.cn-south-1.myhuaweicloud.com/ecfb29bec27c79ff4fc9f94a20be3e10.min
	// 方法：[_0x2f76de(0x165,']^%c')](_0x1ddfea,_0x85c5d,_0x25553c)
	// 在以上文件方法种即可追踪到 CryptoJS 加密的相关参数
	data, err := x.fuckCryptoEncode("a67e9a3a85049339", "86ad9b37cc9f5b9501b3cecc7dc6377c", findStr)
	if err != nil {
		return "", err
	}

	var iframeUrl = "util.SimpleRegEx(string(buf), `<iframe id=\"play_iframe\" border=\"0\" src=\"(\\S+?)\" width=\"100%\"`)"

	buf, err = x.httpWrapper.Get(fmt.Sprintf("%s?data=%s", x.getIframeApiUrl(iframeUrl), url.QueryEscape(data)))
	if err != nil {
		return "", err
	}
	//log.Println("[RepBuff]", string(buf))

	// 解密入口：[_0x2f76de(0x3f2,'lA0#')](_0x18e43b,_0x22739d,_0x3780eb)
	// 需要将该方法转为可读JS，然后使用goja桥接
	resp, err := x.fuckCryptoJSDecode("a67e9a3a85049339", "86ad9b37cc9f5b9501b3cecc7dc6377c", string(buf))
	if err != nil {
		return "", err
	}

	var result = gjson.Parse(resp)

	if !result.Get("data.url").Exists() {
		return "", errors.New("解析JSON失败：" + resp)
	}

	return result.Get("data.url").String(), nil
}

func (x *FiveMovie) getIframeApiUrl(requestUrl string) string {
	log.Println("[IframeRequestUrl]", requestUrl)
	var apiUrl = "https://player.ddzyku.com:3653/get_url_v2"
	return apiUrl
}

func (x *FiveMovie) handleVideoType(v model.Video) model.Video {
	v.Type = "hls"
	return v
}

// 筛选网络请求，找到特定地址（可能是播放地址）返回
func (x *FiveMovie) fiveParseVideoUrl(requestUrl string) string {
	var findUrl string

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			//log.Println("[network.EventRequestWillBeSent]", ev.Type, ev.Request.URL)
			if util.StringInList(ev.Type.String(), []string{"Stylesheet", "Image", "Font"}) {
				ev.Request.URL = ""
			}
			if util.StringInList(util.HandleHost(ev.Request.URL), util.FiveVideoHost) {
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

// 认证页面模拟点击一下
func (x *FiveMovie) fiveGuardClick(requestUrl string) string {
	var findUrl string

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			//log.Println("[network.EventRequestWillBeSent]", ev.Type, ev.Request.URL)
			if util.StringInList(ev.Type.String(), []string{"Stylesheet", "Image", "Font"}) {
				ev.Request.URL = ""
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

	//var res []byte
	err := chromedp.Run(ctx,
		chromedp.Tasks{
			network.Enable(),
			chromedp.Navigate(requestUrl),
			chromedp.WaitVisible("#access"),
			chromedp.Click("#access"),
			chromedp.WaitVisible(".module-heading-search"),
			//chromedp.FullScreenshot(&res, 90),
		},
	)

	if err != nil {
		log.Println("[chromedp.Run.Error]", err.Error())
	}
	//if err := os.WriteFile("fullScreenshot.png", res, fs.ModePerm); err != nil {
	//	log.Fatal(err)
	//}

	return findUrl
}

func (x *FiveMovie) checkFiveWaf(requestUrl string, responseBody []byte) []byte {
	// <script src="/_guard/html.js?js=click_html"></script>
	if !bytes.Contains(responseBody, []byte("_guard/html.js?js=click_html")) {
		return responseBody
	}

	x.fiveGuardClick(requestUrl)

	return nil
}

func (x *FiveMovie) fuckCryptoEncode(key, iv, data string) (string, error) {
	var scriptBuff = append(
		util.FileReadAllBuf(filepath.Join(util.AppPath(), "app/js/crypto-js.min.js")),
		util.FileReadAllBuf(filepath.Join(util.AppPath(), "app/js/fuck-crypto-bridge.js"))...,
	)
	vm := goja.New()
	_, err := vm.RunString(string(scriptBuff))
	if err != nil {
		log.Println("[LoadGojaError]", err.Error())
		return "", err
	}

	var fuckCryptoEncode func(key, iv, data string) string
	err = vm.ExportTo(vm.Get("fuckCryptoEncode"), &fuckCryptoEncode)
	if err != nil {
		log.Println("[ExportGojaFnError]", err.Error())
		return "", err
	}

	var result = fuckCryptoEncode(key, iv, data)

	return result, nil
}

func (x *FiveMovie) fuckCryptoJSDecode(key, iv, data string) (string, error) {
	var scriptBuff = append(
		util.FileReadAllBuf(filepath.Join(util.AppPath(), "app/js/crypto-js.min.js")),
		util.FileReadAllBuf(filepath.Join(util.AppPath(), "app/js/fuck-crypto-bridge.js"))...,
	)
	vm := goja.New()
	_, err := vm.RunString(string(scriptBuff))
	if err != nil {
		log.Println("[LoadGojaError]", err.Error())
		return "", err
	}

	var fuckCryptoDecode func(key, iv, data string) string
	err = vm.ExportTo(vm.Get("fuckCryptoDecode"), &fuckCryptoDecode)
	if err != nil {
		log.Println("[ExportGojaFnError]", err.Error())
		return "", err
	}

	var result = fuckCryptoDecode(key, iv, data)

	return result, nil
}

func (x *FiveMovie) JA3Request(requestUrl string) (cycletls.Response, error) {
	// 竟然跳不过去！！！！！！！！
	client := cycletls.Init()
	response, err := client.Do(requestUrl, cycletls.Options{
		//Body: "",
		Ja3: "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,27-45-23-5-16-0-65037-51-18-13-43-10-35-17513-11-65281,25497-29-23-24,0",
		//UserAgent: "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:87.0) Gecko/20100101 Firefox/87.0",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		Headers: map[string]string{
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
			"Accept-Encoding": "gzip, deflate, br, zstd",
			"Accept-Language": "zh,en-GB;q=0.9,en-US;q=0.8,en;q=0.7,zh-CN;q=0.6",
			"Cache-Control":   "no-cache",
			"Connection":      "close",
			"Host":            util.HandleHost(requestUrl),
		},
	}, "GET")

	if err != nil {
		log.Print("[JA3RequestError]" + err.Error())
	}
	return response, err
}
