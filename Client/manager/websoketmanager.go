package manager

import (
	"awesomeProject/ws"
	"sync"
)

// WebSocketManager 管理全局的 WebSocket 连接
type WebSocketManager struct {
	Client *ws.WebSocketClient
	Mu     sync.Mutex
}

var wsManager *WebSocketManager

// 获取 WebSocketManager 单例

func GetWebSocketManager() *WebSocketManager {
	if wsManager == nil {
		wsManager = &WebSocketManager{}
	}
	return wsManager
}
