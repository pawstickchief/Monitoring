package crond

import (
	"Client/datetype"
	"Client/mode"
	"Client/setting"
	"Client/ws"
	"encoding/json"
	"fmt"
	"log"
)

func ReceiveTaskFile() (interface{}, error) {
	wsManager := ws.GetWebSocketManager()
	address := fmt.Sprintf("http://%s:%d%s", setting.Conf.ServerConfig.Ip, setting.Conf.ServerConfig.Port, setting.Conf.ServerConfig.DownloadApi)

	// 检查连接是否存在
	if wsManager.Client == nil || wsManager.Client.Conn == nil {
		log.Println("WebSocket connection is not available")
		return nil, fmt.Errorf("WebSocket connection is not available")
	}

	// 创建消息请求结构体实例
	request := datetype.TaskRequest{
		RequestId: setting.Conf.ClientIp,
		TaskType:  "get_task", // 自定义消息内容
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

	// 提取 tasklist 数据
	fileList, ok := responseData["files"].([]interface{})
	if !ok {
		log.Println("No tasklist found in the response or tasklist format invalid")
		return nil, fmt.Errorf("invalid tasklist in response")
	}

	// 创建一个字符串切片来存储 taskID
	var fileIds []string

	// 遍历 taskList，将每个元素转换为字符串
	for _, task := range fileList {
		taskID, ok := task.(string)
		if !ok {
			log.Println("Invalid task_id format in tasklist")
			return nil, fmt.Errorf("invalid task_id format in tasklist")
		}
		fileIds = append(fileIds, taskID)
	}
	err = mode.DownloadFile(fileIds, address, setting.Conf.WorkDir)
	if err != nil {
		return nil, err
	}

	return responseData, err
}
