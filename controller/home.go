package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	go_websocket "github.com/lixiang4u/go-websocket"
	"log"
	"net/http"
	"time"
)

type HomeController struct {
}

// 演示默认路由
func (p HomeController) Index(c *gin.Context) {
	c.String(http.StatusOK, "nothing here!")
}

func (p HomeController) Hello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"/path": "/hello", "time": time.Now().String()})
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
	var d = gin.H{"event": data["event"], "client_id": clientId, "timestamp": time.Now().Unix()}
	b, _ := json.MarshalIndent(d, "", "	")
	_ = ws.WriteMessage(messageType, b)
	return true
}

func (p HomeController) Play(c *gin.Context) {
	var clientId = c.PostForm("client_id")
	var id = c.PostForm("id")

	var d = gin.H{
		"event":     "play",
		"client_id": clientId,
		"video":     movieVideoById(id),
		"timestamp": time.Now().Unix(),
	}
	b, _ := json.MarshalIndent(d, "", "	")

	log.Println("[debug]", clientId, string(b))

	go_websocket.WSendMessage(clientId, websocket.TextMessage, b)
	c.JSON(http.StatusOK, gin.H{})
}
