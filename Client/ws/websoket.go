package ws

import (
	"awesomeProject/setting"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
	"time"
)

// WebSocket 客户端结构

type WebSocketClient struct {
	Conn       *websocket.Conn
	Token      string
	ExpiresAt  time.Time
	ServerAddr string
	ClientIP   string
}

// 连接 WebSocket 服务端

func ConnectWebSocketAndRequestToken(serverAddr, clientIP string) (*WebSocketClient, error) {
	u := fmt.Sprintf("ws://%s/wsclient", serverAddr)
	log.Printf("Connecting to %s", u)

	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	conn, _, err := dialer.Dial(u, nil)
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

// 发送消息

// CommunicateWithServer 方法实现
func (client *WebSocketClient) CommunicateWithServer(requesters string, message string) (interface{}, error) {
	authenticatedMessage := map[string]string{
		"Token": client.Token,
		"type":  requesters,
		"Msg":   message,
	}

	authenticatedMessageJSON, err := json.Marshal(authenticatedMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message with token: %v", err)
	}

	err = client.Conn.WriteMessage(websocket.TextMessage, authenticatedMessageJSON)
	if err != nil {
		log.Println("Send message failed, attempting to reconnect...")
		if reconnectErr := client.Reconnect(); reconnectErr != nil {
			return nil, fmt.Errorf("failed to reconnect after message send failure: %v", reconnectErr)
		}

		// 重连成功后再尝试重新发送消息
		err = client.Conn.WriteMessage(websocket.TextMessage, authenticatedMessageJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to send message after reconnect: %v", err)
		}
	}

	// 读取服务器的响应
	_, response, err := client.Conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read server response: %v", err)
	}

	// 打印或处理服务器响应
	log.Printf("Server response: %s", string(response))
	data := string(response)
	return data, nil
}

// 心跳检测

func (client *WebSocketClient) Heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			log.Println("Heartbeat failed, attempting to reconnect...")
			if reconnectErr := client.Reconnect(); reconnectErr != nil {
				log.Println("Reconnection failed:", reconnectErr)
				return
			}
		}
	}
}

// 重连机制

func (client *WebSocketClient) Reconnect() error {
	serverAddress := setting.Conf.ServerConfig.Ip + ":" + strconv.Itoa(setting.Conf.ServerConfig.Port)
	var attempt int
	maxTotalAttempts := setting.Conf.WebSocket.MaxTotalReconnectAttempts // 最大重连次数
	reconnectInterval := time.Duration(setting.Conf.WebSocket.ReconnectInterval) * time.Second

	for attempt = 0; attempt < maxTotalAttempts; attempt++ {
		// 尝试重连
		newClient, err := ConnectWebSocketAndRequestToken(serverAddress, client.ClientIP)
		if err != nil {
			log.Printf("Reconnect attempt %d/%d failed: %v", attempt+1, maxTotalAttempts, err)
			// 指数回退机制：每次重连失败后延迟的时间成倍增加
			time.Sleep(reconnectInterval)
			reconnectInterval *= 2 // 每次失败后延迟时间加倍
			continue
		}

		// 更新连接信息
		client.Conn = newClient.Conn
		client.Token = newClient.Token
		client.ExpiresAt = newClient.ExpiresAt

		log.Println("Reconnection successful!")
		return nil // 重连成功，返回 nil
	}

	// 超过最大重连次数，返回错误信息
	errMsg := fmt.Sprintf("Exceeded maximum reconnect attempts (%d). Stopping reconnection attempts.", maxTotalAttempts)
	log.Println(errMsg)
	return fmt.Errorf(errMsg) // 返回错误信息
}
