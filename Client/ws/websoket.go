package ws

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

// WebSocketClient 结构体
type WebSocketClient struct {
	Conn       *websocket.Conn
	Token      string
	ExpiresAt  time.Time
	ServerAddr string
	ClientIP   string
}

// WebSocketManager 结构体，管理 WebSocket 连接
type WebSocketManager struct {
	Client *WebSocketClient
	Mu     sync.Mutex
}

var WSManager *WebSocketManager

// GetWebSocketManager 获取 WebSocketManager 单例
func GetWebSocketManager() *WebSocketManager {
	if WSManager == nil {
		WSManager = &WebSocketManager{}
	}
	return WSManager
}

// ConnectWebSocketAndRequestToken 连接 WebSocket 并请求 Token
func ConnectWebSocketAndRequestToken(serverAddr, clientIP string) (*WebSocketClient, error) {
	u := fmt.Sprintf("ws://%s/wsclient", serverAddr)
	log.Printf("Connecting to %s", u)

	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket server: %v", err)
	}

	client := &WebSocketClient{Conn: conn, ServerAddr: serverAddr, ClientIP: clientIP}

	request := map[string]string{
		"type":      "request_token",
		"client_ip": clientIP,
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %v", err)
	}

	err = client.Conn.WriteMessage(websocket.TextMessage, requestJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to send token request: %v", err)
	}

	_, response, err := client.Conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %v", err)
	}

	var tokenResponse map[string]string
	err = json.Unmarshal(response, &tokenResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token response: %v", err)
	}

	client.Token = tokenResponse["Token"]
	expiresAt, err := time.Parse(time.RFC3339, tokenResponse["ExpiresAt"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse expiresAt time: %v", err)
	}
	client.ExpiresAt = expiresAt

	log.Printf("Received Token: %s, ExpiresAt: %s", client.Token, client.ExpiresAt)

	return client, nil
}

// Heartbeat 函数，定时发送心跳包
func Heartbeat(client *WebSocketClient) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			log.Println("Heartbeat failed, attempting to reconnect...")
			if reconnectErr := Reconnect(client); reconnectErr != nil {
				log.Println("Reconnection failed:", reconnectErr)
				return
			}
		}
	}
}
func StartWebSocketClient(serverAddr, clientIP string) (*WebSocketClient, error) {
	client, err := ConnectWebSocketAndRequestToken(serverAddr, clientIP)
	if err != nil {
		return nil, err
	}
	go Heartbeat(client)
	return client, nil
}

// Reconnect 函数，尝试重新连接
func Reconnect(client *WebSocketClient) error {
	serverAddress := client.ServerAddr
	var attempt int
	maxTotalAttempts := 3 // 假设最大重连次数为 3
	reconnectInterval := 5 * time.Second

	for attempt = 0; attempt < maxTotalAttempts; attempt++ {
		newClient, err := ConnectWebSocketAndRequestToken(serverAddress, client.ClientIP)
		if err != nil {
			log.Printf("Reconnect attempt %d/%d failed: %v", attempt+1, maxTotalAttempts, err)
			time.Sleep(reconnectInterval)
			reconnectInterval *= 2 // 每次失败后延迟时间加倍
			continue
		}

		client.Conn = newClient.Conn
		client.Token = newClient.Token
		client.ExpiresAt = newClient.ExpiresAt

		log.Println("Reconnection successful!")
		return nil // 重连成功，返回 nil
	}

	return fmt.Errorf("failed to reconnect after %d attempts", maxTotalAttempts)
}

// CommunicateWithServer 与服务器通信
func CommunicateWithServer(client *WebSocketClient, requesters string, message interface{}) (interface{}, error) {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %v", err)
	}

	authenticatedMessage := map[string]string{
		"Token": client.Token,
		"type":  requesters,
		"Msg":   string(messageJSON),
	}

	authenticatedMessageJSON, err := json.Marshal(authenticatedMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal authenticated message: %v", err)
	}

	err = client.Conn.WriteMessage(websocket.TextMessage, authenticatedMessageJSON)
	if err != nil {
		log.Println("Send message failed, attempting to reconnect...")
		if reconnectErr := Reconnect(client); reconnectErr != nil {
			return nil, fmt.Errorf("failed to reconnect after message send failure: %v", reconnectErr)
		}

		err = client.Conn.WriteMessage(websocket.TextMessage, authenticatedMessageJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to send message after reconnect: %v", err)
		}
	}

	_, response, err := client.Conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read server response: %v", err)
	}

	log.Printf("Server response: %s", string(response))
	return string(response), nil
}
