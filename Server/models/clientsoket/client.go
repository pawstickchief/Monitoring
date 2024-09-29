package clientsoket

import (
	"github.com/gorilla/websocket"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type Client struct {
	IP         string
	Authorized bool
	AuthToken  string
	Connection *websocket.Conn
}
type JobMgr struct {
	Kv     clientv3.KV
	Lease  clientv3.Lease
	Clinet *clientv3.Client
}
type TokenInfo struct {
	Token         string
	ExpiresAt     string
	ExpiresAtTime time.Time // 添加 ExpiresAtTime 字段
}
