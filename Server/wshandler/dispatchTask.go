package wshandler

import (
	"Server/common"
	"Server/dao/task" // 确认导入路径
	"Server/dao/task/mysqloption"
	"Server/models/tasktype"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"log"
	"strings"
	"time"
)

// DispatchTaskHandler 处理任务下发的请求
type DispatchTaskHandler struct {
	TaskManager *task.Manager // 任务管理器
	Db          *sqlx.DB
}

// NewDispatchTaskHandler 构造函数，用于初始化 DispatchTaskHandler
func NewDispatchTaskHandler(tm *task.Manager, db *sqlx.DB) *DispatchTaskHandler {
	return &DispatchTaskHandler{
		TaskManager: tm,
		Db:          db,
	}
}

func (h *DispatchTaskHandler) HandleMessage(conn *websocket.Conn, msg map[string]interface{}) error {
	common.ClientsMutex.Lock()
	defer common.ClientsMutex.Unlock()

	// 获取 Base64 编码的消息
	base64Msg, ok := msg["Msg"].(string)
	if !ok {
		return errors.New("Msg 字段无效")
	}
	base64Msg = strings.Trim(base64Msg, `"`)
	decodedMsg, err := base64.StdEncoding.DecodeString(base64Msg)
	if err != nil {
		return fmt.Errorf("解码 Base64 字符串失败: %v", err)
	}

	var parsedMsg map[string]interface{}
	err = json.Unmarshal(decodedMsg, &parsedMsg)
	if err != nil {
		return fmt.Errorf("解析解码后的 JSON 消息失败: %v", err)
	}

	taskType, ok := parsedMsg["task_type"].(string)
	if !ok {
		return errors.New("task_type 无效")
	}
	var response map[string]interface{}

	switch taskType {
	case "select":
		client, ok := parsedMsg["request_id"].(string)
		if !ok {
			return errors.New("request_id 无效")
		}
		taskinfo, err := h.TaskManager.GetTasksByClientIP(client, "inactive")
		if err != nil {
			zap.L().Info("任务获取失败", zap.Error(err))
			return fmt.Errorf("任务获取失败: %v", err)
		}
		response = map[string]interface{}{
			"request_id": client,
			"task_list":  taskinfo,
		}
	case "get_task":
		client, ok := parsedMsg["request_id"].(string)
		if !ok {
			return errors.New("request_id 无效")
		}
		taskinfo, err := h.TaskManager.GetTasksByClientIP(client, "inactive")
		if err != nil {
			zap.L().Info("任务获取失败", zap.Error(err))
		}
		var fileContents []string
		for _, taskfile := range taskinfo {
			keyPrefix := fmt.Sprintf("/tasksfile/%s", taskfile)
			taskFiles, err := h.TaskManager.GetTasks(keyPrefix)
			if err != nil {
				zap.L().Info("任务文件获取失败", zap.Error(err))
				continue
			}
			for _, taskFile := range taskFiles {
				fileContents = append(fileContents, string(taskFile.Value))
			}
		}
		response = map[string]interface{}{
			"request_id": client,
			"files":      fileContents,
		}
	case "update_status":
		taskID, ok := parsedMsg["task_id"].(string)
		if !ok {
			return errors.New("task_id 无效")
		}
		newStatus, ok := parsedMsg["task_status"].(string)
		if !ok {
			return errors.New("new_status 无效")
		}
		err := h.TaskManager.UpdateTaskStatus(taskID, newStatus)
		if err != nil {
			return fmt.Errorf("更新任务状态失败: %v", err)
		}
		response = map[string]interface{}{
			"task_id": taskID,
			"status":  "任务状态更新成功",
		}
	case "query_task":
		taskidlist, ok := parsedMsg["task_id"].([]interface{})
		if !ok {
			return errors.New("task_id 无效")
		}
		taskIDs := make([]string, len(taskidlist))
		for i, id := range taskidlist {
			taskID, ok := id.(string)
			if !ok {
				return errors.New("task_id 列表中的元素无效")
			}
			taskIDs[i] = taskID
		}
		taskRecords, err := mysqloption.GetTaskRecordsByTaskIDs(h.Db, taskIDs)
		if err != nil {
			return fmt.Errorf("批量查询任务记录失败: %v", err)
		}
		response = map[string]interface{}{
			"tasks_info": taskRecords,
		}
	case "task_log_accepted":
		// 解析 client_task_log 字段
		taskLogData, ok := parsedMsg["client_task_log"].(map[string]interface{})
		if !ok {
			return errors.New("client_task_log 无效或缺失")
		}

		// 解析 task_id 为字符串
		taskID, ok := taskLogData["task_id"].(string)
		if !ok || taskID == "" {
			return errors.New("task_id 无效或缺失")
		}

		// 解析 client_ip
		clientIP, ok := taskLogData["client_ip"].(string)
		if !ok || clientIP == "" {
			return errors.New("client_ip 无效或缺失")
		}

		// 解析 output 字段
		output, ok := taskLogData["output"].(string)
		if !ok {
			return errors.New("output 无效或缺失")
		}

		// 解析时间字段
		executionTime := time.Now() // 默认为当前时间
		if execTimeStr, ok := taskLogData["execution_time"].(string); ok {
			if parsedTime, err := time.Parse(time.RFC3339, execTimeStr); err == nil {
				executionTime = parsedTime
			} else {
				log.Printf("execution_time 解析失败，使用当前时间: %v", err)
			}
		}

		completionTime := time.Now() // 默认为当前时间
		if completionTimeStr, ok := taskLogData["completion_time"].(string); ok {
			if parsedTime, err := time.Parse(time.RFC3339, completionTimeStr); err == nil {
				completionTime = parsedTime
			} else {
				log.Printf("completion_time 解析失败，使用当前时间: %v", err)
			}
		}

		// 创建 ClientTaskLog 实例并填充数据
		taskLog := tasktype.ClientTaskLog{
			ClientIP:       clientIP,
			TaskID:         taskID, // 直接使用字符串类型的 task_id
			ExecutionTime:  executionTime,
			Output:         output,
			CompletionTime: completionTime,
			Remarks:        taskLogData["remarks"].(string),
		}
		fmt.Println("客户端日志", taskLog)

		// 将任务日志插入数据库
		logID, err := mysqloption.InsertTaskLog(h.Db, &taskLog)
		if err != nil {
			log.Printf("插入任务日志失败: %v", err)
			return err
		}

		// 返回成功响应
		response = map[string]interface{}{
			"log_id": logID,
			"status": "任务日志已接受并插入",
		}

	default:
		return fmt.Errorf("未知的任务类型: %s", taskType)
	}

	err = common.SendJSONResponse(conn, response)
	if err != nil {
		return fmt.Errorf("发送响应失败: %v", err)
	}

	return nil
}
