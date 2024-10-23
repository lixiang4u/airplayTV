package cmd

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/lixiang4u/airplayTV/controller"
	"github.com/lixiang4u/airplayTV/util"
	go_websocket "github.com/lixiang4u/go-websocket"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"time"
)

var httpServerCmd = &cobra.Command{
	Use:   "serve",
	Short: "start http server",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(fmt.Sprintf("[AppPath] %s", util.AppPath()))

		go Clock()

		//_ = autotls.Run(NewRouter(), "tv.artools.cc")
		//log.Println(viper.GetString("app.addr"))
		_ = NewRouter().Run(viper.GetString("app.addr"))

	},
}

func init() {
	rootCmd.AddCommand(httpServerCmd)
}

func Clock() {
	defer func() { recover() }()

	t := time.NewTicker(time.Second * 86400) // 一天后删除缓存
	for {
		select {
		case <-t.C:
			err := os.RemoveAll(fmt.Sprintf("%s/app/cache/colly/", util.AppPath()))
			log.Println("[time.Ticker]", err)
		}
	}
}

// 初始化websocket
func NewRouterW() go_websocket.WSWrapper {
	homeController := new(controller.HomeController)
	var ws = go_websocket.WSWrapper{}
	ws.Config.Debug = true

	ws.On("info", homeController.InfoW)           //注册列表数据查询
	ws.On("list", homeController.ListW)           //注册列表数据查询
	ws.On("broadcast", homeController.BroadcastW) //注册广播消息

	return ws
}

// 新建路由表
func NewRouter() *gin.Engine {
	r := gin.Default()
	r.Use(gin.Recovery())

	// 使用session中间件
	r.Use(sessions.Sessions("airplayTV", cookie.NewStore([]byte(viper.GetString("app.secret")))))
	r.Use(util.SetGINLoggerFormat())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "HEAD"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	ws := NewRouterW()

	log.Println("[p]", fmt.Sprintf("%s/app/view/**/*", util.AppPath()))
	r.LoadHTMLGlob(fmt.Sprintf("%s/app/view/**/*", util.AppPath()))

	r.Static("/html", "./app/public/")
	r.Static("/upload", "./app/upload/")
	r.Static("/static", "./app/static/")
	r.Static("/m3u8", "./app/m3u8/")

	homeController := new(controller.HomeController)
	videoController := new(controller.VideoController)
	m3u8Controller := new(controller.M3u8Controller)
	m3u8Controller.Init()

	r.GET("/", homeController.Index) // 默认首页

	// 统一api
	r.GET("/api/env/predict", homeController.EnvPredict)
	r.GET("/api/video/search", videoController.Search)
	r.GET("/api/video/tag/:tagName", videoController.ListByTag)
	r.GET("/api/video/detail/:id", videoController.Detail)      // 视频详细信息
	r.GET("/api/video/source/:id", videoController.Source)      // 视频播放信息
	r.GET("/api/video/tag", videoController.ListByTagV2)        // 支持非路径参数
	r.GET("/api/video/detail", videoController.DetailV2)        // 支持非路径参数
	r.GET("/api/video/source", videoController.SourceV2)        // 支持非路径参数
	r.GET("/api/video/airplay", videoController.Airplay)        // 支持非路径参数
	r.GET("/api/video/controls", videoController.VideoControls) // 远程遥控功能
	r.GET("/api/video/cors", videoController.VideoVideoCORS)    // 处理CORS问题 https://api.czspp.com:81/m3/27f4dc5a9b1663195094/ixDR95MU2KkzEZzolGWs0v7FumR9yxWHUhLe7Ea_EgNhL5vaIvS_8BgDvCme9Cl-159akCDxdEgpXAmWMjAO-XIHyn86mudjDSYTmRmApys.m3u8
	r.GET("/api/ws", func(context *gin.Context) {
		ws.Run(context.Writer, context.Request, nil)
	})

	r.GET("/tesla/fullscreen", homeController.FullScreen)
	r.GET("/api/m3u8p", m3u8Controller.Proxy)
	r.HEAD("/api/m3u8p", m3u8Controller.Proxy)

	r.GET("/ws", func(context *gin.Context) {
		ws.Run(context.Writer, context.Request, nil)
	})

	return r
}
