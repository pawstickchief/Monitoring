package route

import (
	"awesomeProject/manager"
	"awesomeProject/setting"
	"awesomeProject/ws"
	"log"
	"strconv"
	"time"
)

// 初始化 WebSocket 连接

func InitWebSocket() {
	wsManager := manager.GetWebSocketManager()

	wsManager.Mu.Lock()
	defer wsManager.Mu.Unlock()

	serverAddress := setting.Conf.ServerConfig.Ip + ":" + strconv.Itoa(setting.Conf.ServerConfig.Port)

	var attempt int

	for attempt = 0; attempt < 3; attempt++ {
		client, err := ws.ConnectWebSocketAndRequestToken(serverAddress, setting.Conf.ClientIp)
		if err != nil {
			log.Printf("Failed to connect, attempt %d/%d: %v", attempt+1, 3, err)
			time.Sleep(10 * time.Second) // 等待重连间隔
			continue
		}

		wsManager.Client = client
		log.Println("WebSocket connection established successfully")
		break
	}
}
