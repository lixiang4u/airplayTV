package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lixiang4u/ShotTv-api/controller"
	go_websocket "github.com/lixiang4u/go-websocket"
)

func main() {

	//_ = autotls.Run(router.NewRouter(), "example1.com")
	_ = NewRouter().Run(":8089")

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

	r.GET("/", new(controller.HomeController).Index)      // 默认首页
	r.GET("/hello", new(controller.HomeController).Hello) // 测试页
	r.GET("/ws", func(context *gin.Context) {
		ws.Run(context.Writer, context.Request, nil)
	})

	return r
}
