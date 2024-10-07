package ws

import (
	"Client/datetype"
	"Client/mode"
	"Client/setting"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func ensureConnection(wsManager *WebSocketManager) error {
	if wsManager == nil {
		return fmt.Errorf("WebSocketManager is nil")
	}
	if wsManager.Client == nil || wsManager.Client.Conn == nil {
		log.Println("WebSocket connection is not available")
		return fmt.Errorf("WebSocket connection is not available")
	}
	return nil
}

// Receive 从 WebSocket 服务器接收任务列表
func Receive(wsManager *WebSocketManager) (interface{}, error) {
	// 检查 WebSocket 连接是否存在
	if err := ensureConnection(wsManager); err != nil {
		return nil, err
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
	response, err := CommunicateWithServer(wsManager.Client, "task_request", requestData)
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
func ReceiveTaskFile(wsManager *WebSocketManager) (interface{}, error) {
	// 构建下载地址
	address := fmt.Sprintf("http://%s:%d%s", setting.Conf.ServerConfig.Ip, setting.Conf.ServerConfig.Port, setting.Conf.ServerConfig.DownloadApi)

	if err := ensureConnection(wsManager); err != nil {
		return nil, err
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
	response, err := CommunicateWithServer(wsManager.Client, "task_request", requestData)
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

	// 提取文件列表数据
	fileList, ok := responseData["files"].([]interface{})
	if !ok {
		log.Println("No tasklist found in the response or tasklist format invalid")
		return nil, fmt.Errorf("invalid tasklist in response")
	}

	// 创建一个字符串切片来存储文件ID
	var fileIds []string

	// 遍历文件列表，将每个元素转换为字符串
	for _, task := range fileList {
		taskID, ok := task.(string)
		if !ok {
			log.Println("Invalid task_id format in tasklist")
			return nil, fmt.Errorf("invalid task_id format in tasklist")
		}
		fileIds = append(fileIds, taskID)
	}

	// 下载任务文件
	err = mode.DownloadFile(fileIds, address, setting.Conf.WorkDir)
	if err != nil {
		return nil, err
	}

	return responseData, err
}
func TaskLogPut(wsManager *WebSocketManager, tasklog datetype.ClientTaskLog) (map[string]interface{}, error) {
	// 构建上传地址
	address := fmt.Sprintf("http://%s:%d%s", setting.Conf.ServerConfig.Ip, setting.Conf.ServerConfig.Port, setting.Conf.ServerConfig.UploadApi)
	if err := ensureConnection(wsManager); err != nil {
		return nil, err
	}
	// 构建请求体
	request := datetype.TaskLogRequest{
		RequestId:     setting.Conf.ClientIp,
		TaskType:      "task_log_accepted",
		ClientTaskLog: tasklog,
	}
	fmt.Println("发送的日志数据", request)

	// 将结构体序列化为 JSON
	requestData, err := json.Marshal(request)
	if err != nil {
		log.Printf("Failed to marshal request data: %v", err)
		return nil, err
	}

	// 使用 tasklog 的 Output 字段作为日志文件的文件名
	logFile := tasklog.Output

	// 检查日志文件是否存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("日志文件 %s 不存在", logFile)
	}

	// 上传文件
	err = mode.UploadFiles([]string{logFile}, address) // 调整为传入文件路径
	if err != nil {
		log.Printf("Failed to upload log file: %v\n", err)
		return nil, err
	}

	log.Printf("Log file %s uploaded successfully\n", logFile)

	// 发送消息并接收响应
	response, err := CommunicateWithServer(wsManager.Client, "task_request", requestData)
	if err != nil {
		log.Printf("Failed to send data: %v", err)
		return nil, err
	}

	// 解析服务端返回的 JSON 数据，提取 file_id
	var responseData struct {
		Files []struct {
			FileID   int64  `json:"file_id"`
			FileName string `json:"file_name"`
		} `json:"files"`
		Status string `json:"status"`
	}
	err = json.Unmarshal([]byte(response.(string)), &responseData)
	if err != nil {
		log.Printf("Failed to unmarshal response: %v", err)
		return nil, err
	}

	// 确认上传状态为 success
	if responseData.Status != "success" {
		return nil, fmt.Errorf("文件上传失败, 响应状态: %s", responseData.Status)
	}

	// 提取第一个文件的 file_id（假设只有一个文件）
	if len(responseData.Files) == 0 {
		return nil, fmt.Errorf("响应中没有 file_id")
	}
	fileID := responseData.Files[0].FileID

	// 返回上传的文件ID信息
	processedTasks := map[string]interface{}{
		"file_id": fileID,
		"status":  "文件上传成功",
	}

	return processedTasks, nil
}
func TaskInfoGet(wsManager *WebSocketManager, taskIDs []string) (map[string]interface{}, error) {
	// 构建请求体
	request := datetype.TaskRequest{
		RequestId: setting.Conf.ClientIp,
		TaskType:  "query_task",
		TaskId:    taskIDs,
	}
	if err := ensureConnection(wsManager); err != nil {
		return nil, err
	}

	// 将结构体序列化为 JSON
	requestData, err := json.Marshal(request)
	if err != nil {
		log.Printf("Failed to marshal request data: %v", err)
		return nil, err
	}

	// 发送消息并接收响应
	response, err := CommunicateWithServer(wsManager.Client, "task_request", requestData)
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
func UpdateTaskStatus(wsManager *WebSocketManager, taskID string, taskStatus string) (interface{}, error) {
	// 构建请求体
	request := datetype.TaskStatus{
		RequestId:  setting.Conf.ClientIp,
		TaskType:   "update_status",
		TaskId:     taskID,
		TaskStatus: taskStatus, // 自定义消息内容
	}
	if err := ensureConnection(wsManager); err != nil {
		return nil, err
	}
	// 将结构体序列化为 JSON
	requestData, err := json.Marshal(request)
	if err != nil {
		log.Printf("Failed to marshal request data: %v", err)
		return nil, err
	}

	// 发送消息并接收响应
	response, err := CommunicateWithServer(wsManager.Client, "task_request", requestData)
	if err != nil {
		log.Printf("Failed to send data: %v", err)
		return nil, err
	}

	return response, err
}
func handleTaskMessage(action string, serverResponse map[string]interface{}) {
	taskID := serverResponse["task_id"].(string)
	crondExpression, _ := serverResponse["crond_expression"].(string) // 如果没有可能是空字符串
	scriptPath, _ := serverResponse["script_path"].(string)           // 同上

	switch action {
	case "add":
		err := WSManager.TaskManager.AddTask(taskID, crondExpression, scriptPath)
		if err != nil {
			log.Printf("添加任务 %s 失败: %v\n", taskID, err)
		} else {
			log.Printf("任务 %s 添加成功\n", taskID)
		}
	case "stop":
		err := WSManager.TaskManager.StopTask(taskID)
		if err != nil {
			log.Printf("停止任务 %s 失败: %v\n", taskID, err)
		} else {
			log.Printf("任务 %s 已停止\n", taskID)
		}
	default:
		log.Printf("未知操作: %s\n", action)
	}
}
func handleGeneralResponse(response map[string]interface{}) {
	log.Printf("Received general response: %v", response)
	// 这里可以进一步解析 response 内容并处理
	// 例如，打印查询结果或更新某些状态
}
