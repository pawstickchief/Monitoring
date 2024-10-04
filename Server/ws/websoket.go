package ws

import (
	"Server/common"
	"Server/dao/task"
	"Server/wshandler"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"log"
	"sync"
	"time"
)

// 客户端 WebSocket 结构

type WebSocketClient struct {
	Conn          *websocket.Conn
	Authorization string
	ExpiresAt     time.Time
}

var (
	Clients      = make(map[*websocket.Conn]*WebSocketClient) // 存储连接的客户端
	ClientsMutex = sync.Mutex{}                               // 保护 clients 访问的互斥锁
)

// 初始化处理器

func InitHandlers(taskManager *task.Manager, db *sqlx.DB) {
	common.RegisterHandler("ping", &wshandler.PingHandler{})
	common.RegisterHandler("update", &wshandler.UpdateHandler{})
	common.RegisterHandler("request_token", &wshandler.TokenHandler{})
	common.RegisterHandler("demo", &wshandler.DemoHandle{})
	common.RegisterHandler("connection_status", &wshandler.ConnectionStatusHandler{})
	common.RegisterHandler("task_request", &wshandler.DispatchTaskHandler{TaskManager: taskManager, Db: db})
}

// WebSocketHandler 处理 WebSocket 连接

func WebsocketHandler(c *gin.Context) {
	conn, err := common.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()
	clientIP := c.ClientIP() // 获取客户端的 IP 地址
	client := &common.WebSocketClient{
		Conn:      conn,
		ClientIP:  clientIP,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	common.ClientsMutex.Lock()
	common.Clients[conn] = client
	common.ClientsMutex.Unlock()

	defer func() {
		common.ClientsMutex.Lock()
		delete(common.Clients, conn)
		common.ClientsMutex.Unlock()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read message failed: %v", err)
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to parse JSON: %v", err)
			continue
		}

		msgType, ok := msg["type"].(string)
		if !ok {
			log.Printf("Invalid message format: missing 'type'")
			continue
		}

		handler, exists := common.GetHandler(msgType)
		if !exists {
			log.Printf("No handler found for message type: %s", msgType)
			continue
		}

		if err := handler.HandleMessage(conn, msg); err != nil {
			log.Printf("Error handling message of type %s: %v", msgType, err)
		}
	}
}
