package controller

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	nnM3u8Url = "https://www.nunuyy2.org/url.php"
	nnPlayUrl = "https://www.nunuyy2.org/dianying/%s.html" //https://www.nunuyy2.org/dianying/101451.html
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
