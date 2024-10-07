package common

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

// WebSocketClient 定义了客户端的结构
type WebSocketClient struct {
	Conn          *websocket.Conn
	ClientIP      string
	Authorization string
	ExpiresAt     time.Time
	ClientID      string // 新增一个客户端ID，用于唯一标识客户端
}
type WebSocketManager struct {
	Clients      map[*websocket.Conn]*WebSocketClient // 管理所有客户端
	ClientsMutex sync.Mutex                           // 锁保护并发读写
}

// 全局客户端管理
var (
	Clients      = make(map[*websocket.Conn]*WebSocketClient)
	ClientsMutex = sync.Mutex{}
)

// WebSocket 升级器，允许跨域

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 消息处理器接口，处理不同类型的消息
// NewWebSocketManager 创建并返回一个新的 WebSocketManager

func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		Clients: make(map[*websocket.Conn]*WebSocketClient),
	}
}

type MessageHandler interface {
	HandleMessage(conn *websocket.Conn, msg map[string]interface{}) error
}

// GetClients 返回所有客户端连接
func (manager *WebSocketManager) GetClients() map[*websocket.Conn]*WebSocketClient {
	manager.ClientsMutex.Lock()
	defer manager.ClientsMutex.Unlock()
	return manager.Clients
}

// 处理器注册表，用于存储不同消息类型的处理器
var handlerRegistry = struct {
	sync.RWMutex
	handlers map[string]MessageHandler
}{
	handlers: make(map[string]MessageHandler),
}

// RegisterHandler 注册一个消息处理器
func RegisterHandler(msgType string, handler MessageHandler) {
	handlerRegistry.Lock()
	defer handlerRegistry.Unlock()
	handlerRegistry.handlers[msgType] = handler
}

// GetHandler 获取一个消息处理器
func GetHandler(msgType string) (MessageHandler, bool) {
	handlerRegistry.RLock()
	defer handlerRegistry.RUnlock()
	handler, exists := handlerRegistry.handlers[msgType]
	return handler, exists
}

// AddClient 添加客户端到全局管理中
func AddClient(conn *websocket.Conn, client *WebSocketClient) {
	ClientsMutex.Lock()
	defer ClientsMutex.Unlock()
	Clients[conn] = client
}

// RemoveClient 从全局管理中删除客户端
func RemoveClient(conn *websocket.Conn) {
	ClientsMutex.Lock()
	defer ClientsMutex.Unlock()
	delete(Clients, conn)
}

// SendJSONResponse 发送 JSON 格式的数据到 WebSocket 连接
func SendJSONResponse(conn *websocket.Conn, message map[string]interface{}) error {
	// 序列化消息为 JSON 格式
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return err
	}

	// 发送 JSON 消息到 WebSocket 连接
	err = conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		log.Printf("Error sending message: %v", err)
		return err
	}

	return nil
}

// BroadcastTaskMessage 向所有客户端广播任务信息
func BroadcastTaskMessage(clients map[*websocket.Conn]*WebSocketClient, taskMessage map[string]interface{}) {
	ClientsMutex.Lock()
	defer ClientsMutex.Unlock()

	for conn, client := range clients {
		// 在消息中添加目标客户端的 IP
		taskMessage["client_ip"] = client.ClientIP
		err := conn.WriteJSON(taskMessage)
		if err != nil {
			log.Printf("Failed to broadcast to client %s: %v", client.ClientIP, err)
			conn.Close()
			delete(clients, conn)
		}
	}
}
