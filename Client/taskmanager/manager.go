package taskmanager

import (
	"Client/ws"
	"log"
)

type Task struct {
	TaskID  string `json:"task_id"`
	Content string `json:"content"`
	Status  string `json:"status"` // inactive, active, completed 等状态
}

type TaskManager struct {
	TaskList  map[string]Task // 使用 map 存储任务ID对应的任务
	WSManager *ws.WebSocketManager
}

func (tm *TaskManager) AddTask(taskID string, crondExpression string, scriptPath string) error {
	// 检查任务是否已经存在
	if _, exists := tm.TaskList[taskID]; exists {
		return nil // 任务已存在，不需要添加
	}

	// 添加新任务到 TaskList
	tm.TaskList[taskID] = Task{
		TaskID:  taskID,
		Content: scriptPath,
		Status:  "active",
	}

	log.Printf("任务 %s 已成功添加，crond 表达式: %s\n", taskID, crondExpression)
	return nil
}

// StopTask 实现 TaskManager 接口中的 StopTask 方法
func (tm *TaskManager) StopTask(taskID string) error {
	// 检查任务是否存在
	task, exists := tm.TaskList[taskID]
	if !exists {
		log.Printf("任务 %s 不存在", taskID)
		return nil
	}

	// 更新任务状态为 "stopped"
	task.Status = "stopped"
	tm.TaskList[taskID] = task
	log.Printf("任务 %s 已停止", taskID)
	return nil
}

// InitTask 从服务器获取任务并初始化任务列表
func (tm *TaskManager) InitTask() {
	// 调用 crond 中的函数，从服务器获取任务
	taskIDs, err := ws.Receive(tm.WSManager)
	if err != nil {
		log.Fatalf("获取任务失败: %v", err)
	}

	// 将接收到的任务添加到 TaskList 中，初始状态设置为 inactive
	for _, taskID := range taskIDs.([]string) {
		tm.TaskList[taskID] = Task{
			TaskID: taskID,
			Status: "active", // 默认将状态设置为 active
		}
	}

	log.Println("任务初始化完成:", tm.TaskList)
}

// UpdateTaskStatus 更新任务的状态
func (tm *TaskManager) UpdateTaskStatus(taskID, status string) {
	// 检查任务是否存在
	task, exists := tm.TaskList[taskID]
	if !exists {
		log.Printf("任务 %s 不存在", taskID)
		return
	}

	// 更新任务状态
	task.Status = status
	tm.TaskList[taskID] = task
	log.Printf("任务 %s 的状态已更新为 %s", taskID, status)
}

// GetTask 获取任务
func (tm *TaskManager) GetTask(taskID string) *Task {
	task, exists := tm.TaskList[taskID]
	if !exists {
		log.Printf("任务 %s 不存在", taskID)
		return nil
	}
	return &task
}
