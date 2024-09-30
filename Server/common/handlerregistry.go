package common

import (
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type WebSocketClient struct {
	Conn          *websocket.Conn
	ClientIP      string
	Authorization string
	ExpiresAt     time.Time
}

var (
	Clients      = make(map[*websocket.Conn]*WebSocketClient)
	ClientsMutex = sync.Mutex{}
)
