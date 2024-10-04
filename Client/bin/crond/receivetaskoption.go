package crond

import (
	"Client/datetype"
	"Client/setting"
	"Client/ws"
	"encoding/json"
	"fmt"
	"log"
)

func TaskInfoGet(taskid []string) (map[string]interface{}, error) {
	wsManager := ws.GetWebSocketManager()

	// 构建请求体
	request := datetype.TaskRequest{
		RequestId: setting.Conf.ClientIp,
		TaskType:  "query_task",
		TaskId:    taskid,
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
	err = json.Unmarshal([]byte(response.(string)), &responseData)
	if err != nil {
		log.Printf("Failed to unmarshal response: %v", err)
		return nil, err
	}

	// 检查并提取 tasks_info 数据
	tasksInfo, ok := responseData["tasks_info"].([]interface{})
	if !ok {
		log.Println("tasks_info 格式不正确，无法执行")
		return nil, fmt.Errorf("invalid tasks_info format in response")
	}

	// 创建一个用于存储预处理后任务详情的 map
	processedTasks := make(map[string]interface{})

	// 遍历 tasks_info 列表，提取每个任务的相关信息
	for _, task := range tasksInfo {
		taskInfo, ok := task.(map[string]interface{})
		if !ok {
			log.Println("任务详情格式不正确，跳过执行")
			continue
		}

		// 提取 task_id 和其他信息
		taskID, ok := taskInfo["task_id"].(string)
		if !ok {
			log.Println("任务 ID 不存在或格式不正确，跳过执行")
			continue
		}

		// 将每个任务的详情存储到 processedTasks 中，以 task_id 作为 key
		processedTasks[taskID] = taskInfo
	}

	// 返回预处理后的任务详情
	return processedTasks, nil
}
