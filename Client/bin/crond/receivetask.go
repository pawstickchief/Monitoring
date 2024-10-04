package crond

import (
	"Client/datetype"
	"Client/setting"

	"Client/ws"
	"encoding/json"
	"fmt"
	"log"
)

func Receive() (interface{}, error) {
	wsManager := ws.GetWebSocketManager()

	// 检查 WebSocket 连接是否存在
	if wsManager.Client == nil || wsManager.Client.Conn == nil {
		log.Println("WebSocket connection is not available")
		return nil, fmt.Errorf("WebSocket connection is not available")
	}

	// 创建消息请求结构体实例
	request := datetype.TaskRequest{
		RequestId: setting.Conf.ClientIp,
		TaskType:  "select", // 自定义消息内容
	}

	// 将结构体序列化为 JSON
	requestData, err := json.Marshal(request)
	if err != nil {
		log.Printf("Failed to marshal request data: %v", err)
		return nil, err
	}

	// 发送消息并接收响应
	response, err := ws.CommunicateWithServer(wsManager.Client, "task_request", requestData)
	if err != nil {
		log.Printf("Failed to send data: %v", err)
		return nil, err
	}

	// 解析服务端返回的数据
	var responseData map[string]interface{}
	err = json.Unmarshal([]byte(response.(string)), &responseData) // 解析返回的字符串为 JSON
	if err != nil {
		log.Printf("Failed to unmarshal response: %v", err)
		return nil, err
	}

	// 提取 task_list 数据，如果不存在则返回空数组
	taskList, ok := responseData["task_list"].([]interface{})
	if !ok || taskList == nil {
		log.Println("No task_list found in the response or task_list format invalid, returning empty list")
		return []string{}, nil // 返回空列表，不触发错误
	}

	// 创建一个字符串切片来存储 taskID
	var taskIds []string

	// 遍历 taskList，将每个元素转换为字符串
	for _, task := range taskList {
		taskID, ok := task.(string)
		if !ok {
			log.Println("Invalid task_id format in task_list")
			return nil, fmt.Errorf("invalid task_id format in task_list")
		}
		taskIds = append(taskIds, taskID)
	}

	// 返回提取到的任务列表
	return taskIds, nil
}
