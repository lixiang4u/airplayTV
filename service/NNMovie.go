package service

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	nnHost      = "https://www.huibangpaint.com/"
	nnM3u8Url   = "https://nnyy.in/url.php"
	nnPlayUrl   = "https://www.huibangpaint.com/vodplay/%s.html"
	nnSearchUrl = "https://www.huibangpaint.com/vodsearch/%s----------%d---.html"
	nnTagUrl    = "https://www.huibangpaint.com/vodtype/1-%d.html" //https://www.nunuyy2.org/dianying/index_3.html
	nnDetailUrl = "https://www.huibangpaint.com/voddetail/%s.html"
)

type NNMovie struct{ Movie }

func (x NNMovie) ListByTag(tagName, page string) model.Pager {
	return x.nnListByTag("dianying", page)
}

func (x NNMovie) Search(search, page string) model.Pager {
	return x.nnListBySearch(search, page)
}

func (x NNMovie) Detail(id string) model.MovieInfo {
	return x.nnVideoDetail(id)
}

func (x NNMovie) Source(sid, vid string) model.Video {
	return x.nnVideoSource(sid, vid)
}

//========================================================================
//==============================实际业务处理逻辑============================
//========================================================================

func (x NNMovie) nnListBySearch(search, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 10 // 每页10条

	c := x.Movie.NewColly()

	c.OnHTML("#searchList .thumb", func(element *colly.HTMLElement) {
		name := element.ChildAttr("a", "title")
		tmpUrl := element.ChildAttr("a", "href")
		thumb := element.ChildAttr("a", "data-original")
		tag := element.ChildText(".pic-text")

		pager.List = append(pager.List, model.MovieInfo{
			Id:    util.SimpleRegEx(tmpUrl, `(\d+)`),
			Name:  name,
			Thumb: thumb,
			Url:   util.FillUrlHost(tmpUrl, nnHost),
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

	err := c.Visit(fmt.Sprintf(nnSearchUrl, search, handleNNPageNumber(page)))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return pager
}

func (x NNMovie) nnListByTag(tagName, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 48 // 每页48条

	c := x.Movie.NewColly()

	c.OnHTML("ul.myui-vodlist .myui-vodlist__box", func(element *colly.HTMLElement) {
		name := element.ChildAttr("a.myui-vodlist__thumb", "title")
		tmpUrl := element.ChildAttr("a.myui-vodlist__thumb", "href")
		thumb := element.ChildAttr("a.myui-vodlist__thumb", "data-original")
		tag := element.ChildText(".pic-text")

		pager.List = append(pager.List, model.MovieInfo{
			Id:    util.SimpleRegEx(tmpUrl, `(\d+)`),
			Name:  name,
			Thumb: util.FillUrlHost(thumb, nnHost),
			Url:   util.FillUrlHost(tmpUrl, nnHost),
			Tag:   tag,
		})
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	regEx := regexp.MustCompile(`/vodtype/(\d+)-(\d+).html`)

	c.OnHTML(".myui-page", func(element *colly.HTMLElement) {
		element.ForEach("li a", func(i int, element *colly.HTMLElement) {
			tmpList := regEx.FindStringSubmatch(element.Attr("href"))
			if len(tmpList) != 3 {
				return
			}
			n, _ := strconv.Atoi(tmpList[2])
			if n > pager.Total {
				pager.Total = n
			}
		})

		pager.Current, _ = strconv.Atoi(page)
	})

	tmpUrl := fmt.Sprintf(nnTagUrl, util.HandlePageNumber(page))
	err := c.Visit(tmpUrl)
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return pager
}

// 根据id获取视频播放列表信息
func (x NNMovie) nnVideoDetail(id string) model.MovieInfo {
	var info = model.MovieInfo{Id: id}
	var sourceMap = make(map[string]string, 0)

	c := x.Movie.NewColly()
	c.OnHTML(".myui-content__thumb", func(element *colly.HTMLElement) {
		info.Thumb = util.FillUrlHost(element.ChildAttr("a img", "data-original"), nnHost)
		info.Name = element.ChildAttr("a", "title")
	})
	c.OnHTML("meta[name=description]", func(element *colly.HTMLElement) {
		info.Intro = element.Attr("content")
	})
	c.OnHTML(".myui-panel_hd .nav-tabs", func(element *colly.HTMLElement) {
		element.ForEach("li a", func(i int, element *colly.HTMLElement) {
			sourceMap[strings.TrimLeft(element.Attr("href"), "#")] = element.Text
		})
	})
	c.OnHTML(".tab-content", func(element *colly.HTMLElement) {
		element.ForEach(".tab-pane", func(groupIndex int, element *colly.HTMLElement) {
			var sourceId = element.Attr("id")
			groupName, ok := sourceMap[sourceId]
			if !ok {
				groupName = fmt.Sprintf("来源%d", groupIndex+1)
			}
			element.ForEach("li a", func(i int, element *colly.HTMLElement) {
				info.Links = append(info.Links, model.Link{
					Id:    util.SimpleRegEx(element.Attr("href"), `(\d+-\d+-\d+)`),
					Name:  element.Text,
					Url:   util.FillUrlHost(element.Attr("href"), nnHost),
					Group: groupName,
				})
			})
		})
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf(nnDetailUrl, id))
	if err != nil {
		log.Println("[visit.error]", err.Error())
	}

	return info
}

func (x NNMovie) nnVideoSource(sid, vid string) model.Video {
	var video = model.Video{Id: sid, Source: sid}

	//获取基础信息
	c := x.Movie.NewColly()
	c.OnHTML(".myui-player__data", func(element *colly.HTMLElement) {
		video.Name = element.ChildText(".text-fff")
		video.Thumb = ""
	})
	c.OnHTML(".embed-responsive", func(element *colly.HTMLElement) {
		video.Source = util.SimpleRegEx(element.Text, `"url":"(\S+?)",`)
		video.Source = strings.ReplaceAll(video.Source, "\\/", "/")
		video.Type = util.GuessVideoType(video.Source)

		video.Url = HandleSrcM3U8FileToLocal(sid, video.Source, x.Movie.IsCache)
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL.String())
	})

	err := c.Visit(fmt.Sprintf(nnPlayUrl, sid))
	if err != nil {
		log.Println("[visit.error]", err.Error())
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
func handleNNVideoPlayLinks(idPathUrl string) (links []model.Link, err error) {
	var res []interface{}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	err = chromedp.Run(
		ctx,
		chromedp.Navigate(fmt.Sprintf(nnPlayUrl, idPathUrl)),
		chromedp.Evaluate(`urls;`, &res),
	)
	if err != nil && !strings.Contains(err.Error(), "net::ERR_ABORTED") {
		// Note: Ignoring the net::ERR_ABORTED page error is essential here
		// since downloads will cause this error to be emitted, although the
		// download will still succeed.
		log.Println("[network.error]", err)
		return
	}
	if res == nil {
		log.Println("empty....")
		return
	}

	var counter = 1
	for idx0, tmpList := range res {
		var tmpGroup = ""
		for idx1, tmpUrl := range tmpList.([]interface{}) {
			if tmpUrl == nil {
				continue
			}
			if idx1 == 0 {
				tmpGroup = fmt.Sprintf("group_%d", idx0+1)
			}
			links = append(links, model.Link{
				Id:    tmpUrl.(string),
				Name:  fmt.Sprintf("资源%d", idx1+1),
				Url:   tmpUrl.(string),
				Group: tmpGroup,
			})
			counter++
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

func nnHandleUrlToId(tmpUrl string) string {
	tmpUrl = strings.TrimRight(tmpUrl, ".html")
	tmpUrl = strings.ReplaceAll(tmpUrl, "/", "-")
	return strings.TrimLeft(tmpUrl, "-")
}

func handleNNImageUrl(tmpUrl string) string {
	if util.IsHttpUrl(tmpUrl) == true {
		return tmpUrl
	}
	return fmt.Sprintf("%s/%s", util.HandleHost(nnM3u8Url), strings.TrimLeft(tmpUrl, "/"))
}
