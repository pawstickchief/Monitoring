package crond

import (
	"Client/datetype"
	"Client/setting"
	"Client/ws"
	"encoding/json"
	"log"
)

func UpdateTaskStatus(TaskId string, TaskStatus string) (interface{}, error) {
	wsManager := ws.GetWebSocketManager()
	request := datetype.TaskStatus{
		RequestId:  setting.Conf.ClientIp,
		TaskType:   "update_status",
		TaskId:     TaskId,
		TaskStatus: TaskStatus, // 自定义消息内容
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

	return response, err
}
