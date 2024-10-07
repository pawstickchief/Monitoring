package clientoption

import (
	"github.com/jmoiron/sqlx"
	"log"
	"time"
)

// InsertClientConnection 插入客户端连接信息到数据库
// InsertOrUpdateClientConnection 插入或更新客户端连接信息到数据库

func InsertOrUpdateClientConnection(db *sqlx.DB, clientIP, authCode string, expires time.Time) error {
	// 插入或更新客户端连接信息
	_, err := db.Exec(`
		INSERT INTO client_server_connections (client_ip, connection_time, auth_code, auth_expiration)
		VALUES (?, NOW(), ?, ?)
		ON DUPLICATE KEY UPDATE
		connection_time = NOW(),
		auth_code = VALUES(auth_code),
		auth_expiration = VALUES(auth_expiration)
	`, clientIP, authCode, expires.Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Printf("Failed to insert or update client connection: %v", err)
		return err
	}
	return nil
}

// UpdateClientConnection 更新客户端连接的过期时间
func UpdateClientConnection(db *sqlx.DB, clientIP, authCode string) error {
	// 更新客户端连接信息
	_, err := db.Exec(`
		UPDATE client_server_connections 
		SET auth_expiration = DATE_ADD(NOW(), INTERVAL 1 DAY) 
		WHERE client_ip = ? AND auth_code = ?
	`, clientIP, authCode)
	if err != nil {
		log.Printf("Failed to update client connection: %v", err)
		return err
	}
	return nil
}

// DeleteClientConnection 删除客户端连接
func DeleteClientConnection(db *sqlx.DB, clientIP string) error {
	// 删除客户端连接信息
	_, err := db.Exec(`DELETE FROM client_server_connections WHERE client_ip = ?`, clientIP)
	if err != nil {
		log.Printf("Failed to delete client connection: %v", err)
		return err
	}
	return nil
}
