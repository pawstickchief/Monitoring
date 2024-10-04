package websoket

import "time"

type ClientServerConnection struct {
	ConnectionID   int       `json:"connection_id" db:"connection_id"`     // 连接 ID
	ClientIP       string    `json:"client_ip" db:"client_ip"`             // 客户端 IP
	ConnectionTime time.Time `json:"connection_time" db:"connection_time"` // 连接时间
	AuthExpiration time.Time `json:"auth_expiration" db:"auth_expiration"` // 授权过期时间
	AuthCode       string    `json:"auth_code" db:"auth_code"`             // 授权码
	Remarks        string    `json:"remarks" db:"remarks"`                 // 备注
}
