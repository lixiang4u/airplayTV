package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/lixiang4u/airplayTV/util"
	"io/ioutil"
	"log"
	"time"
)

func init() {
	util.LoadConfig()
}

func main() {
	AAAAAA()
	//AWs()
	//util.InitLog()
	//
	//cmd.Execute()

}

// https://github.com/chromedp/chromedp/pull/990
func AWs() {
	var devToolWsUrl string
	var title string

	flag.StringVar(&devToolWsUrl, "devtools-ws-url", "wss://chrome.browserless.io", "DevTools Websocket URL")
	flag.Parse()

	actxt, cancelActxt := chromedp.NewRemoteAllocator(context.Background(), devToolWsUrl)
	defer cancelActxt()

	ctxt, cancelCtxt := chromedp.NewContext(actxt) // create new tab
	defer cancelCtxt()                             // close tab afterwards

	if err := chromedp.Run(ctxt,
		chromedp.Navigate("http://ws.artools.cc/"),
		chromedp.Title(&title),
	); err != nil {
		log.Fatalf("Failed getting body of duckduckgo.com: %v", err)
	}

	log.Println("Got title of:", title)

}

func A() {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	//opts := append(chromedp.DefaultExecAllocatorOptions[:],
	//	chromedp.Flag("headless", false),
	//	chromedp.Flag("disable-gpu", false),
	//	chromedp.Flag("enable-automation", false),
	//	chromedp.Flag("disable-extensions", false),
	//)

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			//log.Println("[network.EventRequestWillBeSent]", ev.Request.URL)
			log.Println("[ev.Request.Headers]", ev.Request.Headers)
		case *network.EventResponseReceived:
			//log.Println("[network.EventResponseReceived]", ev.RequestID.String(), ev.Response.URL)
			if ev.Response.MimeType == "image/jpeg" || ev.Response.MimeType == "image/png" {
				break
			}

			// ev.Response.URL == "https://www.czspp.com/reyingzhong"
			if true {

				log.Println("[返回数据]", ev.RequestID, ev.Response.MimeType, ev.Response.URL)
				log.Println("[返回数据Header]", ev.RequestID, ev.Response.Headers)

				//tmpCtx := chromedp.FromContext(ctx)
				//
				//go func() {
				//	body, err := network.GetResponseBody(ev.RequestID).Do(cdp.WithExecutor(ctx, tmpCtx.Target))
				//
				//	if err != nil {
				//		log.Println("[getBodyError]", err.Error())
				//	}
				//	log.Println("[=========================>]", string(body))
				//	if err = os.WriteFile("tmp/"+ev.RequestID.String(), body, 0644); err != nil {
				//		log.Fatal(err)
				//	}
				//}()

				go func() {
					// print response body
					c := chromedp.FromContext(ctx)
					rbp := network.GetResponseBody(ev.RequestID)
					body, err := rbp.Do(cdp.WithExecutor(ctx, c.Target))
					if err != nil {
						fmt.Println("[ev.RequestID]", ev.RequestID, err)
					}
					if err = ioutil.WriteFile("tmp/"+ev.RequestID.String(), body, 0644); err != nil {
						log.Fatal("[ev.RequestID]", ev.RequestID, err)
					}
					if err == nil {
						fmt.Printf("[OKKKKKK]%s\n", ev.RequestID)
					}
				}()
			}
		}
	})

	//ch := chromedp.WaitNewTarget(ctx, func(info *target.Info) bool {
	//	log.Println("[chromedp.WaitNewTarget]", info.URL)
	//	return false
	//})

	err := chromedp.Run(ctx,
		chromedp.Tasks{
			network.Enable(),
			chromedp.Navigate("https://www.czspp.com/reyingzhong"),
			//chromedp.Navigate("https://www.nunuyy5.org/"),

			chromedp.Sleep(time.Second * 20),

			chromedp.Poll("false", chromedp.ByID),
			//chromedp.Sleep(time.Second * 20),
			//
			//chromedp.WaitEnabled("#sdsdsd", chromedp.ByID),
			////chromedp.WaitNotPresent(),
			////chromedp.WaitReady(),
			//chromedp.WaitVisible("#dgfgfgf"),
			//chromedp.Poll(),
			//chromedp.PollFunction(),
		},
	)

	if err != nil {
		log.Println("[chromedp.Run.Error]", err.Error())
	}

	log.Println("[run....]")
	//<-ch

	time.Sleep(time.Minute)
}

func mainA() {
	//to remove defaults
	ctx_, cancel_ := chromedp.NewExecAllocator(context.Background())
	defer cancel_()
	ctx, cancel := chromedp.NewContext(ctx_)
	defer cancel()

	chromedp.ListenTarget(ctx, func(event interface{}) {
		switch ev := event.(type) {
		case *network.EventResponseReceived:
			println(ev.RequestID)
			body, err := network.GetResponseBody(ev.RequestID).Do(ctx)
			if err != nil {
				log.Println("getting body error: ", err)
				return
			}
			if err = ioutil.WriteFile(ev.RequestID.String(), body, 0644); err != nil {
				log.Fatal(err)
			}
		}
	})
	err := chromedp.Run(ctx, tasks())
	if err != nil {
		log.Fatal(err)
	}
}

func tasks() chromedp.Tasks {
	return chromedp.Tasks{
		network.Enable(),
		chromedp.Navigate("https://google.com/"),
		chromedp.Sleep(time.Second * 10),
	}
}

func AAAAAA() {
	// https://555movie.me/vodplay/391558-1-1.html

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	//opts := append(chromedp.DefaultExecAllocatorOptions[:],
	//	chromedp.Flag("headless", false),
	//	chromedp.Flag("disable-gpu", false),
	//	chromedp.Flag("enable-automation", false),
	//	chromedp.Flag("disable-extensions", false),
	//)

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			//log.Println("[network.EventRequestWillBeSent]", ev.Request.URL)
			log.Println("[ev.Request.URL]", ev.Request.URL)
			log.Println("[ev.Request.Headers]", ev.Request.Headers)
		case *network.EventResponseReceived:
			//log.Println("[network.EventResponseReceived]", ev.RequestID.String(), ev.Response.URL)
			if ev.Response.MimeType == "image/jpeg" || ev.Response.MimeType == "image/png" {
				//break
			}

			// ev.Response.URL == "https://www.czspp.com/reyingzhong"
			if true {

				//log.Println("[返回数据]", ev.RequestID, ev.Response.MimeType, ev.Response.URL)
				//log.Println("[返回数据Header]", ev.RequestID, ev.Response.Headers)

				//tmpCtx := chromedp.FromContext(ctx)
				//
				//go func() {
				//	body, err := network.GetResponseBody(ev.RequestID).Do(cdp.WithExecutor(ctx, tmpCtx.Target))
				//
				//	if err != nil {
				//		log.Println("[getBodyError]", err.Error())
				//	}
				//	log.Println("[=========================>]", string(body))
				//	if err = os.WriteFile("tmp/"+ev.RequestID.String(), body, 0644); err != nil {
				//		log.Fatal(err)
				//	}
				//}()

				go func() {
					// print response body
					c := chromedp.FromContext(ctx)
					rbp := network.GetResponseBody(ev.RequestID)
					body, err := rbp.Do(cdp.WithExecutor(ctx, c.Target))
					if err != nil {
						fmt.Println("[ev.RequestID]", ev.RequestID, err)
					}
					if err = ioutil.WriteFile("tmp/"+ev.RequestID.String(), body, 0644); err != nil {
						log.Fatal("[ev.RequestID]", ev.RequestID, err)
					}
					if err == nil {
						fmt.Printf("[OKKKKKK]%s\n", ev.RequestID)
					}
				}()
			}
		}
	})

	//ch := chromedp.WaitNewTarget(ctx, func(info *target.Info) bool {
	//	log.Println("[chromedp.WaitNewTarget]", info.URL)
	//	return false
	//})

	err := chromedp.Run(ctx,
		chromedp.Tasks{
			network.Enable(),
			chromedp.Navigate("http://ws.artools.cc/"),
			//chromedp.Navigate("https://www.nunuyy5.org/"),

			chromedp.Poll("false", chromedp.ByID),

			chromedp.Sleep(time.Second * 20),

			//chromedp.Sleep(time.Second * 20),
			//
			//chromedp.WaitEnabled("#sdsdsd", chromedp.ByID),
			////chromedp.WaitNotPresent(),
			////chromedp.WaitReady(),
			//chromedp.WaitVisible("#dgfgfgf"),
			//chromedp.Poll(),
			//chromedp.PollFunction(),
		},
	)

	if err != nil {
		log.Println("[chromedp.Run.Error]", err.Error())
	}

	log.Println("[run....]")
	//<-ch

	time.Sleep(time.Minute)

}
