package route

import (
	"awesomeProject/bin"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// 初始化 WebSocket 升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有请求，生产环境中应根据需要进行限制
		return true
	},
}

// WebSocket 处理器
func WebsocketHandler(c *gin.Context) {
	// 从 URL 查询参数中获取用户ID（或其他标识）
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// 升级 HTTP 连接到 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.L().Info("Failed to upgrade connection: %v", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade to websocket"})
		return
	}
	defer conn.Close()
	zap.L().Info("WebSocket connection established for user:",
		zap.String("userID", userID),
		zap.Error(err),
	)
	// 创建上下文和取消函数
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 在退出时调用取消函数

	// 用于 ping 输出的通道
	outputChan := make(chan string)
	// 互斥锁，确保不会并发写 WebSocket 连接
	var writeMutex sync.Mutex

	// 启动 Goroutine 逐条发送 ping 输出
	go func() {
		for output := range outputChan {
			// 使用互斥锁保护 WebSocket 写操作
			writeMutex.Lock()
			err := conn.WriteMessage(websocket.TextMessage, []byte(output))
			writeMutex.Unlock()
			if err != nil {
				log.Printf("Write error: %v", err)
				return
			}
		}
	}()

	// 监听客户端发送的消息
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		command := string(message)
		log.Printf("Received command: %s", command)

		// 如果收到取消指令，取消当前 ping 操作
		if strings.TrimSpace(command) == "cancel" {
			cancel()          // 触发取消信号
			writeMutex.Lock() // 确保在取消时没有并发写操作
			conn.WriteMessage(websocket.TextMessage, []byte("Ping operation canceled by client"))
			writeMutex.Unlock()
			continue
		}

		// 解析客户端发送的命令格式，如: "ping www.baidu.com 100"
		parts := strings.Split(command, " ")
		if len(parts) != 3 || parts[0] != "ping" {
			writeMutex.Lock()
			conn.WriteMessage(websocket.TextMessage, []byte("Invalid command format. Use 'ping <address> <count>'"))
			writeMutex.Unlock()
			continue
		}

		address := parts[1]                  // 获取目标地址
		count, err := strconv.Atoi(parts[2]) // 获取次数并转换为整数
		if err != nil || count <= 0 {
			writeMutex.Lock()
			conn.WriteMessage(websocket.TextMessage, []byte("Invalid ping count. It must be a positive integer."))
			writeMutex.Unlock()
			continue
		}

		// 重置上下文，用于新的 ping 操作
		ctx, cancel = context.WithCancel(context.Background())

		// 执行 ping 命令
		go func() {
			err := bin.PingWithCancel(ctx, address, count, outputChan) // 执行 ping，次数由客户端决定
			if err != nil {
				writeMutex.Lock()
				conn.WriteMessage(websocket.TextMessage, []byte("Ping command failed: "+err.Error()))
				writeMutex.Unlock()
			}
		}()
	}
}
