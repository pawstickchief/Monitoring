package ws

import (
	"Server/common"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

// 控制指定客户端的指定任务

func ControlClientTask(c *gin.Context) {
	clientID := c.Param("client_ip") // 从 URL 中获取客户端 ID
	taskID := c.Query("task_id")     // 从查询参数中获取任务 ID
	action := c.Query("action")
	crondExpression := c.Query("crondExpression")
	scriptPath := c.Query("scriptPath")
	if err := SendTaskToClient(clientID, taskID, action, crondExpression, scriptPath); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "任务发送成功"})
}

// SendTaskToClient 发送任务指令到指定客户端
// SendTaskToClient 发送任务指令到指定客户端
func SendTaskToClient(clientIP, taskID, action, crondExpression, scriptPath string) error {
	common.ClientsMutex.Lock()

	// 查找目标客户端
	var targetClient *common.WebSocketClient
	for _, client := range common.Clients {
		fmt.Println(client)
		// 如果 WebSocket 连接的地址匹配，就认为是目标客户端
		if client.Conn.RemoteAddr().String() == clientIP {
			targetClient = client
			break
		}
	}

	// 如果未找到目标客户端，打印当前客户端列表
	if targetClient == nil {
		log.Printf("客户端未找到：IP %s", clientIP)
		log.Println("当前连接的客户端列表：")

		// 打印所有当前连接的客户端
		for _, client := range common.Clients {
			log.Printf("客户端 IP: %s, WebSocket地址: %s, 过期时间: %s",
				client.ClientIP, client.Conn.RemoteAddr().String(), client.ExpiresAt.Format(time.RFC3339))
		}

		common.ClientsMutex.Unlock() // 解锁
		return fmt.Errorf("客户端未找到：IP %s", clientIP)
	}

	conn := targetClient.Conn
	common.ClientsMutex.Unlock() // 提前解锁

	// 创建任务消息
	taskMessage := map[string]interface{}{
		"action":           action,          // add, stop 等操作
		"task_id":          taskID,          // 任务 ID
		"crond_expression": crondExpression, // 定时任务表达式
		"script_path":      scriptPath,      // 脚本路径
	}

	// 将任务消息编码为 JSON 并发送
	err := conn.WriteJSON(taskMessage)
	if err != nil {
		// 记录发送失败的错误信息
		log.Printf("发送任务给客户端 %s 失败: %v", clientIP, err)

		// WebSocket 出现问题后，可以考虑移除该客户端连接
		common.ClientsMutex.Lock()
		delete(common.Clients, conn)
		common.ClientsMutex.Unlock()

		return fmt.Errorf("发送任务给客户端 %s 失败: %v", clientIP, err)
	}

	log.Printf("任务 %s 已成功发送给客户端 %s", taskID, clientIP)
	return nil
}
