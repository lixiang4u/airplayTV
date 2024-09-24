package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lixiang4u/airplayTV/model"
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

// 演示websocket
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
		"ua":          ctx.Request.UserAgent(),
		"is_tv":       util.IsTv(ctx.Request.UserAgent()),
		"source_list": p.getAppSourceList(),
	}
	ctx.JSON(http.StatusOK, data)
}

func (p HomeController) getAppSourceList() []model.AppSource {
	var lst = make([]model.AppSource, 0)
	lst = append(lst, model.AppSource{
		Tag:         "ys",
		Description: "源6(ys)",
	})
	lst = append(lst, model.AppSource{
		Tag:         "xk",
		Description: "源7(xk)",
	})
	lst = append(lst, model.AppSource{
		Tag:         "cz",
		Description: "源1(cz)（废弃/需要显卡模拟绕过CF验证）",
	})
	lst = append(lst, model.AppSource{
		Tag:         "nn",
		Description: "源6(ys)",
	})
	lst = append(lst, model.AppSource{
		Tag:         "my",
		Description: "源6(ys)",
	})
	lst = append(lst, model.AppSource{
		Tag:         "ys",
		Description: "源3(my)",
	})
	lst = append(lst, model.AppSource{
		Tag:         "lv",
		Description: "源4(lv)（如果效果不行请更换其他源）(海外)（片源质量差）",
	})
	lst = append(lst, model.AppSource{
		Tag:         "five",
		Description: "源5(five)",
	})

	return lst
}
