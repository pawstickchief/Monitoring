package common

import (
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

// WebSocket 升级器，允许跨域

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 消息处理器接口

type MessageHandler interface {
	HandleMessage(conn *websocket.Conn, msg map[string]interface{}) error
}

// 处理器注册表
var handlerRegistry = struct {
	sync.RWMutex
	handlers map[string]MessageHandler
}{
	handlers: make(map[string]MessageHandler),
}

// 注册处理器

func RegisterHandler(msgType string, handler MessageHandler) {
	handlerRegistry.Lock()
	defer handlerRegistry.Unlock()
	handlerRegistry.handlers[msgType] = handler
}

// 获取处理器

func GetHandler(msgType string) (MessageHandler, bool) {
	handlerRegistry.RLock()
	defer handlerRegistry.RUnlock()
	handler, exists := handlerRegistry.handlers[msgType]
	return handler, exists
}
