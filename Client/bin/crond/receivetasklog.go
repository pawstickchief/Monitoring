package crond

import (
	"Client/datetype"
	"Client/mode"
	"Client/setting"

	"Client/ws"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// 压缩数据
func compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TaskLogPut(tasklog datetype.ClientTaskLog) (map[string]interface{}, error) {
	wsManager := ws.GetWebSocketManager()
	address := fmt.Sprintf("http://%s:%d%s", setting.Conf.ServerConfig.Ip, setting.Conf.ServerConfig.Port, setting.Conf.ServerConfig.UploadApi)

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
	response, err := ws.CommunicateWithServer(wsManager.Client, "task_request", requestData)
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
