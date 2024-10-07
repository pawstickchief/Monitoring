package wshandler

import (
	"Server/common"
	"Server/controller"
	"Server/dao/clientoption"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"log"
	"time"
)

// Update 消息处理器

type UpdateHandler struct{}

func (h *UpdateHandler) HandleMessage(conn *websocket.Conn, msg map[string]interface{}) error {
	data, ok := msg["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid 'data' field")
	}
	log.Printf("Received update: %v", data)
	return nil
}

// Ping 消息处理器

type PingHandler struct{}

func (h *PingHandler) HandleMessage(conn *websocket.Conn, msg map[string]interface{}) error {
	// 改用 map[string]interface{} 类型
	response := map[string]interface{}{"response": "pong"}
	return common.SendJSONResponse(conn, response)
}

// Token 消息处理器

type TokenHandler struct {
	DB *sqlx.DB
}

func (h *TokenHandler) HandleMessage(conn *websocket.Conn, msg map[string]interface{}) error {
	clientIP, ok := msg["client_ip"].(string)
	if !ok || clientIP == "" {
		return fmt.Errorf("client_ip field missing or invalid")
	}

	// 获取 token 和 expiresAt
	token, expiresAt, err := controller.GetTokenForClientFromController(clientIP)
	if err != nil {
		return fmt.Errorf("failed to get token for client: %v", err)
	}

	// 更新客户端连接信息到数据库
	err = clientoption.InsertOrUpdateClientConnection(h.DB, clientIP, token, expiresAt)
	if err != nil {
		log.Printf("Failed to update client connection for client IP %s: %v", clientIP, err)
		return err
	}

	// 返回 token 和过期时间给客户端
	response := map[string]interface{}{
		"Token":     token,
		"ExpiresAt": expiresAt.Format(time.RFC3339),
	}

	return common.SendJSONResponse(conn, response)
}

type DemoHandle struct{}

func (h *DemoHandle) HandleMessage(conn *websocket.Conn, msg map[string]interface{}) error {
	response := map[string]interface{}{"response": "接收数据"}
	return common.SendJSONResponse(conn, response)
}
