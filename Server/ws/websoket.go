package ws

import (
	"Server/common"
	"Server/controller"
	"Server/dao/clientoption"
	"Server/dao/task"
	"Server/wshandler"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

func InitHandlers(taskManager *task.Manager, db *sqlx.DB, etcd *clientv3.Client) {
	common.RegisterHandler("ping", &wshandler.PingHandler{})
	common.RegisterHandler("update", &wshandler.UpdateHandler{})
	common.RegisterHandler("request_token", &wshandler.TokenHandler{DB: db})
	common.RegisterHandler("demo", &wshandler.DemoHandle{})
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
	db, exists := c.Get("db")
	if !exists {
		log.Println("Database connection not found")
		return
	}
	sqlDB, ok := db.(*sqlx.DB) // 进行类型断言，确保是 *sqlx.DB
	if !ok {
		log.Println("Invalid database connection")
		return
	}
	clientIP := c.ClientIP() // 获取客户端的 IP 地址
	client := &common.WebSocketClient{
		Conn:      conn,
		ClientIP:  clientIP,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// 将客户端添加到 Clients 映射
	common.ClientsMutex.Lock()
	common.Clients[conn] = client
	common.ClientsMutex.Unlock()
	token, expiresAt, _ := controller.GetTokenForClientFromController(clientIP)

	// 插入连接记录到数据库
	err = clientoption.InsertOrUpdateClientConnection(sqlDB, clientIP, token, expiresAt) // 假设 auth_code 传入
	if err != nil {
		log.Printf("Failed to insert client connection: %v", err)
	}

	defer func() {
		// 删除数据库中的客户端记录
		err = clientoption.DeleteClientConnection(sqlDB, clientIP)
		if err != nil {
			log.Printf("Failed to delete client connection: %v", err)
		}

		// 客户端断开时移除连接
		common.ClientsMutex.Lock()
		delete(common.Clients, conn)
		common.ClientsMutex.Unlock()
	}()

	// 监听 WebSocket 消息
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
