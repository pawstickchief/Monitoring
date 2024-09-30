package bin

import (
	"awesomeProject/manager" // 引用 manager 包
	"fmt"
	"log"
)

func SendData() (interface{}, error) {
	// 获取 WebSocketManager
	wsManager := manager.GetWebSocketManager()

	// 检查连接是否存在
	if wsManager.Client == nil || wsManager.Client.Conn == nil {
		log.Println("WebSocket connection is not available")
		return nil, fmt.Errorf("WebSocket connection is not available")
	}

	// 发送消息
	repose, err := wsManager.Client.CommunicateWithServer("connection_status", "Hi")
	if err != nil {
		log.Printf("Failed to send data: %v", err)
	}
	return repose, err
}
