package taskexce

import (
	"Client/mode"
	"Client/setting"
	"Client/taskmanager"
	"Client/ws"
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
)

// TaskExecutor 任务执行器结构体，包含本地的任务管理器和定时调度器
// TaskExecutor 用于管理任务执行和调度
type TaskExecutor struct {
	TM            *taskmanager.TaskManager // 任务管理器
	CronScheduler *cron.Cron               // 定时任务调度器
	taskEntryMap  map[string]cron.EntryID  // 记录 taskID 和 entryID 的映射关系
	wsManager     *ws.WebSocketManager
}

// InitExecutor 初始化任务执行器
func InitExecutor(tm *taskmanager.TaskManager, wsManager *ws.WebSocketManager) *TaskExecutor {
	te := &TaskExecutor{
		TM:            tm,
		CronScheduler: cron.New(cron.WithSeconds()), // 初始化 Cron 支持秒级调度
		wsManager:     wsManager,
	}

	te.CronScheduler.Start() // 启动调度器

	return te
}

// Start 函数调度任务并记录任务的 entryID
func (te *TaskExecutor) Start() {
	te.taskEntryMap = make(map[string]cron.EntryID)

	var activeTaskIDs []string
	var fileIDs []string

	// 展示任务列表
	fmt.Println("当前任务列表:")
	for taskID, task := range te.TM.TaskList {
		fmt.Printf("任务ID: %s, 状态: %s, 内容: %s\n", taskID, task.Status, task.Content)
		if task.Status == "active" {
			fmt.Printf("任务 %s 已激活，开始获取任务详情...\n", taskID)
			activeTaskIDs = append(activeTaskIDs, taskID)
		} else {
			fmt.Printf("任务 %s 状态为 %s，未激活，跳过\n", taskID, task.Status)
		}
	}

	if len(activeTaskIDs) == 0 {
		fmt.Println("没有激活的任务，跳过执行")
		return
	}

	// 获取任务详情
	taskDetails, err := ws.TaskInfoGet(te.wsManager, activeTaskIDs)
	if err != nil {
		log.Printf("获取任务详情失败: %v\n", err)
		return
	}

	for taskID, details := range taskDetails {
		taskInfo, ok := details.(map[string]interface{})
		if !ok {
			log.Printf("任务 %s 的详情格式不正确，无法执行", taskID)
			continue
		}

		// 修正 file_id 类型为 int64
		FileId, ok := taskInfo["file_id"].(float64)
		if !ok {
			log.Printf("任务 %s 缺少 FileId，无法定时执行", taskID)
			continue
		}
		fileIDs = append(fileIDs, fmt.Sprintf("%.0f", FileId))

		crondExpression, ok := taskInfo["crond_expression"].(string)
		if !ok || crondExpression == "" {
			log.Printf("任务 %s 缺少 crond_expression，无法定时执行", taskID)
			continue
		}

		scriptPath, ok := taskInfo["script_path"].(string)
		if !ok || scriptPath == "" {
			log.Printf("任务 %s 缺少 script_path，无法执行", taskID)
			continue
		}

		currentTaskID := taskID
		currentScriptPath := scriptPath

		fmt.Printf("任务 %s 的脚本路径为 %s，开始调度\n", currentTaskID, currentScriptPath)

		taskFunc := func() {
			te.ExecuteTask(currentTaskID, currentScriptPath)
		}

		entryID, err := te.CronScheduler.AddFunc(crondExpression, taskFunc)
		if err != nil {
			log.Printf("无法为任务 %s 添加调度: %v", currentTaskID, err)
			continue
		}

		te.taskEntryMap[currentTaskID] = entryID

		te.CronScheduler.Start()

		entry := te.CronScheduler.Entry(entryID)
		nextRun := entry.Next
		if nextRun.IsZero() {
			log.Printf("无法获取任务 %s 的下次执行时间，可能调度器尚未启动", currentTaskID)
		} else {
			fmt.Printf("任务 %s 已成功调度，crond 表达式: %s，下次执行时间: %v\n", currentTaskID, crondExpression, nextRun)
		}
	}

	// 下载文件
	downloadAddress := fmt.Sprintf("http://%s:%d%s", setting.Conf.ServerConfig.Ip, setting.Conf.ServerConfig.Port, setting.Conf.ServerConfig.DownloadApi)

	err = mode.DownloadFile(fileIDs, downloadAddress, setting.Conf.WorkDir)
	if err != nil {
		log.Printf("文件下载失败: %v\n", err)
		return
	}
}

// StopTask 通过 taskID 来关闭任务
func (te *TaskExecutor) StopTask(taskID string) error {
	entryID, exists := te.taskEntryMap[taskID]
	if !exists {
		return fmt.Errorf("任务 %s 不存在或没有调度", taskID)
	}

	te.CronScheduler.Remove(entryID)
	delete(te.taskEntryMap, taskID)

	fmt.Printf("任务 %s 已被关闭\n", taskID)
	return nil
}

// Stop 停止任务调度器
func (te *TaskExecutor) Stop() {
	te.CronScheduler.Stop() // 停止任务调度器
}

// AddTask 动态添加任务并调度
func (te *TaskExecutor) AddTask(taskID string, crondExpression string, scriptPath string) error {
	// 检查任务是否已经存在
	if _, exists := te.taskEntryMap[taskID]; exists {
		return fmt.Errorf("任务 %s 已存在，无法重复添加", taskID)
	}

	// 定义任务执行函数
	taskFunc := func() {
		te.ExecuteTask(taskID, scriptPath)
	}

	// 添加调度任务
	entryID, err := te.CronScheduler.AddFunc(crondExpression, taskFunc)
	if err != nil {
		return fmt.Errorf("无法为任务 %s 添加调度: %v", taskID, err)
	}

	// 记录 entryID
	te.taskEntryMap[taskID] = entryID

	fmt.Printf("任务 %s 已成功添加，crond 表达式: %s\n", taskID, crondExpression)
	return nil
}
