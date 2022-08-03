package cmd

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lixiang4u/ShotTv-api/controller"
	"github.com/lixiang4u/ShotTv-api/util"
	go_websocket "github.com/lixiang4u/go-websocket"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

var httpServerCmd = &cobra.Command{
	Use:   "serve",
	Short: "start http server",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(fmt.Sprintf("[AppPath] %s", util.AppPath()))

		//_ = autotls.Run(NewRouter(), "tv.artools.cc")
		//log.Println(viper.GetString("app.addr"))
		_ = NewRouter().Run(viper.GetString("app.addr"))

	},
}

func init() {
	rootCmd.AddCommand(httpServerCmd)
}

// 初始化websocket
func NewRouterW() go_websocket.WSWrapper {
	var ws = go_websocket.WSWrapper{}
	ws.Config.Debug = true

	ws.On("info", new(controller.HomeController).InfoW)           //注册列表数据查询
	ws.On("list", new(controller.HomeController).ListW)           //注册列表数据查询
	ws.On("broadcast", new(controller.HomeController).BroadcastW) //注册广播消息

	return ws
}

// 新建路由表
func NewRouter() *gin.Engine {
	r := gin.Default()
	ws := NewRouterW()

	log.Println("[p]", fmt.Sprintf("%s/app/view/**/*", util.AppPath()))
	r.LoadHTMLGlob(fmt.Sprintf("%s/app/view/**/*", util.AppPath()))
	//r.LoadHTMLGlob("D:\\repo\\ShotTv-api\\app\\view\\**\\*")

	r.Static("/html", "./app/public/")
	r.Static("/upload", "./app/upload/")
	r.Static("/static", "./app/static/")

	r.GET("/", new(controller.HomeController).Index)      // 默认首页
	r.GET("/hello", new(controller.HomeController).Hello) // 测试页
	r.POST("/api/play", new(controller.HomeController).Play)
	r.GET("/api/search", new(controller.ResourceController).Search)
	r.GET("/api/tag", new(controller.ResourceController).ListByTag)
	r.GET("/api/tag/:tagName", new(controller.ResourceController).ListByTag)
	r.GET("/api/info/:id", new(controller.ResourceController).Info)
	r.GET("/api/video/:id", new(controller.ResourceController).Video)

	r.GET("/home", new(controller.ResourceController).Home)
	r.GET("/info/:id", new(controller.ResourceController).Info)

	r.GET("/ws", func(context *gin.Context) {
		ws.Run(context.Writer, context.Request, nil)
	})

	return r
}
