package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go-web-app/controller"
	"go.uber.org/zap"
	"time"
)

// 客户端 WebSocket 结构

type WebSocketClient struct {
	Conn          *websocket.Conn
	Authorization string
	ExpiresAt     time.Time
}

// WebSocket 处理器

func WebsocketHandler(c *gin.Context) {
	// 升级 HTTP 连接为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.L().Error("Failed to upgrade connection", zap.Error(err))
		return
	}
	defer conn.Close()

	clientIP := c.ClientIP()

	// 获取或生成 token
	token, expiresAt, err := controller.GetTokenForClientFromController(clientIP)
	if err != nil {
		zap.L().Error("Failed to retrieve token", zap.Error(err))
		return
	}

	// 发送 token 和剩余有效时间给客户端
	if err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Token: %s, ExpiresAt: %s", token, expiresAt.Format(time.RFC3339)))); err != nil {
		zap.L().Error("Failed to send token to client", zap.Error(err))
		return
	}

	// 处理客户端消息
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			zap.L().Error("Read message failed", zap.Error(err))
			break
		}
		zap.L().Info("Received message", zap.String("message", string(message)))
	}
}
