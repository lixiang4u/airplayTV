package controller

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/ShotTv-api/model"
	"github.com/lixiang4u/ShotTv-api/util"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	nnM3u8Url   = "https://www.nunuyy2.org/url.php"
	nnPlayUrl   = "https://www.nunuyy2.org/dianying/%s.html" //https://www.nunuyy2.org/dianying/101451.html
	nnSearchUrl = "https://www.nunuyy2.org/so/%s-%s-%d-.html"
)

// 使用chromedp直接请求页面关联的播放数据m3u8
// 应该可以直接从chromedp拿到m3u8地址，但是没跑通，可以先拿到请求所需的所有上下文，然后http.Post拿数据
func GetNNVideoUrl(id string) (videoUrl string, err error) {

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch ev := v.(type) {
		case *network.EventRequestWillBeSent:
			err := handleNNVideoUrl(ev.Request.URL, ev.Request.PostData, &videoUrl)
			if err != nil {
				panic(err)
			}
		case *network.EventResponseReceived:
			// see: https://github.com/chromedp/chromedp/issues/713#issuecomment-731338643
		case network.EventLoadingFinished:
		case *network.EventLoadingFinished:
			// see: https://github.com/chromedp/chromedp/issues/713#issuecomment-731338643
			// 本来是需要通过network.EventResponseReceived后再到这里注入network.GetResponseBody(reqestId)去拿数据的，奈何跑不通，
			// 只能在network.EventRequestWillBeSent使用http.Post请求拿数据
		}
	})

	err = chromedp.Run(ctx, chromedp.Navigate(fmt.Sprintf(nnPlayUrl, id)))
	if err != nil && !strings.Contains(err.Error(), "net::ERR_ABORTED") {
		// Note: Ignoring the net::ERR_ABORTED page error is essential here
		// since downloads will cause this error to be emitted, although the
		// download will still succeed.
		//log.Fatal(err)
	}
	return
}

func handleNNVideoUrl(requestUrl, postData string, videoUrl *string) error {
	if !strings.HasSuffix(strings.TrimSpace(requestUrl), "url.php") {
		return nil
	}
	log.Println("[nn.request.post]", requestUrl, postData)
	resp, err := http.Post(nnM3u8Url, "application/x-www-form-urlencoded", strings.NewReader(postData))
	if err != nil {
		log.Println("[nn.request.post.error]", err)
		return err
	}
	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println("[nn.request.response.error]", err)
		return err
	}
	var m3u8Url = string(b)
	if strings.Contains(m3u8Url, "script") && strings.Contains(m3u8Url, "body") {
		log.Println("[nn.response.error]", m3u8Url[:50])
		m3u8Url = "解析播放地址失败"
	}
	*videoUrl = m3u8Url
	return nil
}

func GetNNSearchList(search, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 24 // 每页24条

	c := colly.NewCollector(colly.CacheDir(util.GetCollyCacheDir()))

	c.OnHTML(".lists-content li", func(element *colly.HTMLElement) {
		name := element.ChildText("h2 a")
		url := element.ChildAttr("a.thumbnail", "href")
		thumb := element.ChildAttr("img.thumb", "src")
		tag := element.ChildText(".note")

		pager.List = append(pager.List, model.MovieInfo{
			Id:    handleUrlToId(url),
			Name:  name,
			Thumb: thumb,
			Url:   url,
			Tag:   tag,
		})
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	c.OnHTML(".dytop .dy_tit_big", func(element *colly.HTMLElement) {
		element.ForEach("span", func(i int, element *colly.HTMLElement) {
			if i == 0 {
				pager.Total, _ = strconv.Atoi(element.Text)
			}
		})
	})

	c.OnHTML(".pagination", func(element *colly.HTMLElement) {
		currentPageText := element.ChildText(".active span")
		pageIndex := -1
		element.ForEach("li a", func(i int, element *colly.HTMLElement) {
			tmpList := strings.Split(element.Attr("href"), "-")
			if len(tmpList) != 4 {
				return
			}
			n, _ := strconv.Atoi(tmpList[2])
			if n > pageIndex {
				pageIndex = n
				pager.Total = pager.Limit * pageIndex
			}
		})

		pager.Current, _ = strconv.Atoi(currentPageText)
	})

	err := c.Visit(fmt.Sprintf(nnSearchUrl, search, search, handleNNPageNumber(page)))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return pager
}

// 计算页数值，页数从0开始
func handleNNPageNumber(page string) int {
	// https://www.nunuyy2.org/so/僵尸-僵尸-0-.html 第一页
	// https://www.nunuyy2.org/so/僵尸-僵尸-1-.html 第二页
	// https://www.nunuyy2.org/so/僵尸-僵尸-9-.html 第十页

	n, err := strconv.Atoi(page)
	if err != nil {
		return 0
	}
	if n <= 0 {
		return 0
	}
	return n - 1
}

// 根据视频id获取播放列表
func GetNNVideoPlayLinks(id string) (urls []string, err error) {
	var res []interface{}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = chromedp.Run(
		ctx,
		chromedp.Navigate(fmt.Sprintf(nnPlayUrl, id)),
		chromedp.Evaluate(`urls;`, &res),
	)
	if err != nil && !strings.Contains(err.Error(), "net::ERR_ABORTED") {
		// Note: Ignoring the net::ERR_ABORTED page error is essential here
		// since downloads will cause this error to be emitted, although the
		// download will still succeed.
		//log.Fatal(err)
		return
	}
	if res == nil {
		return
	}

	for _, tmpList := range res {
		for _, url := range tmpList.([]interface{}) {
			urls = append(urls, url.(string))
		}
	}

	return
}

func wrapLinks(urls []string) []model.Link {
	var links []model.Link
	for idx, u := range urls {
		links = append(links, model.Link{
			Id:   u,
			Name: fmt.Sprintf("资源%d", idx+1),
			Url:  u,
		})
	}
	return links
}

// 根据id获取视频播放列表信息
func GetNNVideoInfo(id string) model.MovieInfo {
	var info = model.MovieInfo{}

	info.Id = id

	c := colly.NewCollector(colly.CacheDir(util.GetCollyCacheDir()))

	c.OnHTML(".product-header", func(element *colly.HTMLElement) {
		info.Thumb = element.ChildAttr(".thumb", "src")
		info.Name = element.ChildText(".product-title")
		info.Intro = element.ChildText(".product-excerpt span")
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf(nnPlayUrl, id))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}
	urls, err := GetNNVideoPlayLinks(id)

	if err == nil {
		info.Links = wrapLinks(urls)
	}

	return info
}
