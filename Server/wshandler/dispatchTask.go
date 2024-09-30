package wshandler

import (
	"errors"
	"github.com/gorilla/websocket"
	"go-web-app/common"
	"go-web-app/dao/task" // 确认导入路径
	"log"
	"time"
)

// DispatchTaskHandler 处理任务下发的请求
type DispatchTaskHandler struct {
	taskManager *task.Manager // 任务管理器
}

// NewDispatchTaskHandler 构造函数，用于初始化 DispatchTaskHandler
func NewDispatchTaskHandler(tm *task.Manager) *DispatchTaskHandler {
	return &DispatchTaskHandler{
		taskManager: tm,
	}
}

// HandleMessage 处理客户端请求，根据授权码获取任务并下发
func (h *DispatchTaskHandler) HandleMessage(conn *websocket.Conn, msg map[string]interface{}) error {
	common.ClientsMutex.Lock()
	defer common.ClientsMutex.Unlock()

	// 获取请求中的 request_id，用于一对一响应
	requestID, ok := msg["request_id"].(string)
	if !ok {
		return errors.New("request_id 无效")
	}

	// 获取客户端发送的授权码
	authCode, ok := msg["auth_code"].(string)
	if !ok || authCode == "" {
		return errors.New("授权码无效")
	}

	log.Printf("处理客户端请求: request_id=%s, auth_code=%s", requestID, authCode)

	// 初始化返回响应，包含连接数和每个连接的 IP 和连接时间
	response := map[string]interface{}{
		"request_id":       requestID, // 关联请求 ID
		"connection_count": len(common.Clients),
		"clients":          []map[string]interface{}{},
	}

	// 遍历当前连接的客户端信息
	for _, client := range common.Clients {
		clientInfo := map[string]interface{}{
			"client_ip": client.ClientIP,
			"expiresAt": client.ExpiresAt.Format(time.RFC3339),
		}
		response["clients"] = append(response["clients"].([]map[string]interface{}), clientInfo)
	}

	// 调用任务调度器，根据授权码获取任务数据
	fullTasks, err := h.taskManager.GetTasksByAuthCode(authCode)
	if err != nil {
		log.Printf("获取任务失败: %v", err)
		return err
	}

	// 构造任务信息列表
	response["tasks"] = fullTasks

	// 发送响应给客户端，确保请求与答复是一对一的
	err = common.SendJSONResponse(conn, response)
	if err != nil {
		log.Printf("发送响应失败: %v", err)
	}

	return nil
}
