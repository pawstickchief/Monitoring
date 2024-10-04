package wshandler

import (
	"Server/common"
	"Server/controller"
	"fmt"
	"github.com/gorilla/websocket"
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

type TokenHandler struct{}

func (h *TokenHandler) HandleMessage(conn *websocket.Conn, msg map[string]interface{}) error {
	clientIP, ok := msg["client_ip"].(string)
	if !ok || clientIP == "" {
		return fmt.Errorf("client_ip field missing or invalid")
	}

	token, expiresAt, err := controller.GetTokenForClientFromController(clientIP)
	if err != nil {
		return fmt.Errorf("failed to get token for client: %v", err)
	}

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

type ConnectionStatusHandler struct{}

func (h *ConnectionStatusHandler) HandleMessage(conn *websocket.Conn, msg map[string]interface{}) error {
	common.ClientsMutex.Lock()
	defer common.ClientsMutex.Unlock()

	// 返回连接数和每个连接的 IP 和连接时间等信息
	response := map[string]interface{}{
		"connection_count": len(common.Clients),
		"clients":          make([]map[string]interface{}, 0),
	}

	for _, client := range common.Clients {
		clientInfo := map[string]interface{}{
			"client_ip": client.ClientIP,
			"expiresAt": client.ExpiresAt.Format(time.RFC3339),
		}
		response["clients"] = append(response["clients"].([]map[string]interface{}), clientInfo)
	}

	// 发送响应
	return common.SendJSONResponse(conn, response)
}
