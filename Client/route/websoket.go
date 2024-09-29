package route

import (
	"awesomeProject/setting"
	"crypto/tls"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
)

// 客户端 WebSocket 结构

type WebSocketClient struct {
	Conn          *websocket.Conn
	Authorization string
}

// 初始化 WebSocket 连接

func InitWebSocket() {
	serverAddress := setting.Conf.ServerConfig.Ip + ":" + strconv.Itoa(setting.Conf.ServerConfig.Port)
	client, err := ConnectWebSocket(serverAddress)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer client.Conn.Close()

	// 发送客户端IP地址
	if err := client.SendIPAddress(setting.Conf.ClientIp); err != nil {
		log.Fatalf("Failed to send IP address: %v", err)
	}

	// 接收授权码
	if err := client.ReceiveAuthorizationCode(); err != nil {
		log.Fatalf("Failed to receive authorization code: %v", err)
	}

	// 之后可以继续和服务端进行通信
	if err := client.CommunicateWithServer("Some Message"); err != nil {
		log.Fatalf("Failed to communicate with server: %v", err)
	}
}

// 连接 WebSocket 服务端

func ConnectWebSocket(serverAddr string) (*WebSocketClient, error) {
	// 设置 WebSocket 连接的地址
	u := fmt.Sprintf("ws://%s/wsclient", serverAddr)
	log.Printf("Connecting to %s", u)

	// 配置跳过证书验证 (对于 HTTPS)
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// 建立 WebSocket 连接
	conn, _, err := dialer.Dial(u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket server: %v", err)
	}

	return &WebSocketClient{Conn: conn}, nil
}

// 发送 IP 地址给服务端

func (client *WebSocketClient) SendIPAddress(ipAddress string) error {
	err := client.Conn.WriteMessage(websocket.TextMessage, []byte(ipAddress))
	if err != nil {
		return fmt.Errorf("failed to send IP address: %v", err)
	}
	return nil
}

// 接收授权码

func (client *WebSocketClient) ReceiveAuthorizationCode() error {
	_, message, err := client.Conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to receive authorization code: %v", err)
	}
	client.Authorization = string(message)
	log.Printf("Received Authorization Code: %s", client.Authorization)
	return nil
}

// 使用授权码进行通信

func (client *WebSocketClient) CommunicateWithServer(message string) error {
	// 包含授权码的消息发送
	authenticatedMessage := fmt.Sprintf("AuthCode:%s, Msg:%s", client.Authorization, message)
	err := client.Conn.WriteMessage(websocket.TextMessage, []byte(authenticatedMessage))
	if err != nil {
		return fmt.Errorf("failed to send message with auth code: %v", err)
	}
	return nil
}
