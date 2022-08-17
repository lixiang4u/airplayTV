package service

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/ShotTv-api/model"
	"github.com/lixiang4u/ShotTv-api/util"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	nnM3u8Url   = "https://www.nunuyy2.org/url.php"
	nnPlayUrl   = "https://www.nunuyy2.org/dianying/%s.html"
	nnSearchUrl = "https://www.nunuyy2.org/so/%s-%s-%d-.html"
)

type NNMovie struct{}

func (x NNMovie) ListByTag(tagName, page string) model.Pager {
	return nnListBySearch("天", page)
}

func (x NNMovie) Search(search, page string) model.Pager {
	return nnListBySearch(search, page)
}

func (x NNMovie) Detail(id string) model.MovieInfo {
	return nnVideoDetail(id)
}

func (x NNMovie) Source(id string) model.Video {
	log.Println("[ID.......]", id)
	return nnVideoSource(id)
}

//========================================================================
//==============================实际业务处理逻辑============================
//========================================================================

func nnListBySearch(search, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 24 // 每页24条

	c := colly.NewCollector(colly.CacheDir(util.GetCollyCacheDir()))

	c.OnHTML(".lists-content li", func(element *colly.HTMLElement) {
		name := element.ChildText("h2 a")
		url := element.ChildAttr("a.thumbnail", "href")
		thumb := element.ChildAttr("img.thumb", "src")
		tag := element.ChildText(".note")

		pager.List = append(pager.List, model.MovieInfo{
			Id:    util.CZHandleUrlToId(url),
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

// 根据id获取视频播放列表信息
func nnVideoDetail(id string) model.MovieInfo {
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
	urls, err := handleNNVideoPlayLinks(id)

	if err == nil {
		info.Links = wrapLinks(urls)
	}

	return info
}

// 使用chromedp直接请求页面关联的播放数据m3u8
// 应该可以直接从chromedp拿到m3u8地址，但是没跑通，可以先拿到请求所需的所有上下文，然后http.Post拿数据
func nnVideoSource(id string) model.Video {
	var video = model.Video{}

	var v = url.Values{}
	v.Add("url", id)
	//v.Add("sign", strconv.FormatInt(time.Now().Unix(), 10))

	postData, _ := url.QueryUnescape(v.Encode())

	err := handleNNVideoUrl(postData, &video.Url)
	if err != nil {
		video.Name = "[err]" + err.Error()
	}

	return video
}

func handleNNVideoUrl(postData string, videoUrl *string) error {
	log.Println("[nn.request.post]", postData)
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

	log.Println("[nn.response.body]", m3u8Url)

	if strings.Contains(m3u8Url, "script") && strings.Contains(m3u8Url, "body") {
		log.Println("[nn.response.error]", m3u8Url[:50])
		m3u8Url = "解析播放地址失败"
	}
	*videoUrl = m3u8Url
	return nil
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
func handleNNVideoPlayLinks(id string) (urls []string, err error) {
	var res []interface{}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
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
		log.Println("[network]", err)
		return
	}
	if res == nil {
		log.Println("empty....")
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
