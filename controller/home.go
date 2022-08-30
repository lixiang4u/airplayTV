package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lixiang4u/airplayTV/util"
	go_websocket "github.com/lixiang4u/go-websocket"
	"net/http"
	"strings"
	"time"
)

type HomeController struct {
}

// 演示默认路由
func (p HomeController) Index(ctx *gin.Context) {
	_, _ = ctx.Writer.WriteString(strings.TrimSpace(`
<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>AirplayTV api server</title>
</head>
<body>

<h1>AirplayTV!</h1>

</body>
</html>
`))
}

//演示websocket
func (p HomeController) ListW(clientId string, ws *websocket.Conn, messageType int, data map[string]interface{}) bool {
	var d = gin.H{"event": data["event"], "data": go_websocket.WSConnectionList()}
	b, _ := json.MarshalIndent(d, "", "	")
	_ = ws.WriteMessage(messageType, b)
	return true
}

func (p HomeController) BroadcastW(clientId string, ws *websocket.Conn, messageType int, data map[string]interface{}) bool {
	b, _ := json.MarshalIndent(data, "", "	")
	go_websocket.WSBroadcast(clientId, messageType, b)
	return true
}

func (p HomeController) InfoW(clientId string, ws *websocket.Conn, messageType int, data map[string]interface{}) bool {
	var d = gin.H{
		"event":     data["event"],
		"client_id": clientId,
		"timestamp": time.Now().Unix(),
		"msg":       fmt.Sprintf("当前客户端ID: %s", clientId),
	}
	b, _ := json.MarshalIndent(d, "", "	")
	_ = ws.WriteMessage(messageType, b)
	return true
}

func (p HomeController) FullScreen(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "home/fullscreen.html", gin.H{})
}

// 预置环境检测
func (p HomeController) EnvPredict(ctx *gin.Context) {
	var data = gin.H{
		"ua":    ctx.Request.UserAgent(),
		"is_tv": util.IsTv(ctx.Request.UserAgent()),
	}
	ctx.JSON(http.StatusOK, data)
}
