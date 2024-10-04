package taskexce

import (
	"Client/bin/crond"
	"Client/datetype"
	"Client/setting"
	"Client/taskmanager"
	"fmt"
	"github.com/robfig/cron/v3"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// TaskExecutor 任务执行器结构体，包含本地的任务管理器和定时调度器
// TaskExecutor 用于管理任务执行和调度
type TaskExecutor struct {
	TM            *taskmanager.TaskManager // 任务管理器
	CronScheduler *cron.Cron               // 定时任务调度器
	taskEntryMap  map[string]cron.EntryID  // 记录 taskID 和 entryID 的映射关系
}

// InitExecutor 初始化任务执行器
func InitExecutor(tm *taskmanager.TaskManager) *TaskExecutor {
	return &TaskExecutor{
		TM:            tm,
		CronScheduler: cron.New(cron.WithSeconds()), // 初始化 Cron 支持秒级调度
	}
}

// Start 函数调度任务并记录任务的 entryID
func (te *TaskExecutor) Start() {
	te.taskEntryMap = make(map[string]cron.EntryID)

	var activeTaskIDs []string
	for taskID, task := range te.TM.TaskList {
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

	taskDetails, err := crond.TaskInfoGet(activeTaskIDs)
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

func (te *TaskExecutor) ExecuteTask(taskID string, scriptPath string) {
	logFile := te.getLogFilePath(taskID)

	// 打开日志文件
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("无法创建任务 %s 的日志文件: %v\n", taskID, err)
		return
	}
	defer f.Close()

	// 创建日志记录器
	logger := log.New(f, "", log.LstdFlags)

	// 记录任务开始执行
	logger.Printf("开始执行任务 %s，脚本路径: %s\n", taskID, scriptPath)

	// 判断操作系统类型并执行任务
	switch runtime.GOOS {
	case "windows":
		logger.Printf("在 Windows 平台上执行任务 %s，脚本路径: %s\n", taskID, scriptPath)
		te.executeWindowsScript(taskID, scriptPath, logger)
	case "linux":
		logger.Printf("在 Linux 平台上执行任务 %s，脚本路径: %s\n", taskID, scriptPath)
		te.executeLinuxScript(taskID, scriptPath, logger)
	default:
		logger.Printf("不支持的操作系统: %s\n", runtime.GOOS)
		return
	}

	// 更新任务状态为 completed
	te.TM.TaskList[taskID] = taskmanager.Task{
		TaskID: taskID,
		Status: "completed",
	}

	// 创建任务日志结构体，将日志文件名作为 Output 字段
	taskLog := datetype.ClientTaskLog{
		ClientIP:       setting.Conf.ClientIp,
		TaskID:         taskID,
		ExecutionTime:  time.Now(),
		Output:         logFile, // 将日志文件的路径作为任务输出
		CompletionTime: time.Now(),
		Remarks:        "任务执行日志",
	}

	// 发送日志到服务端
	_, err = crond.TaskLogPut(taskLog)
	if err != nil {
		logger.Printf("发送任务 %s 的日志到服务端失败: %v\n", taskID, err)
	}

	// 记录任务完成
	logger.Printf("任务 %s 执行完成\n", taskID)
}

// executeWindowsScript 执行 Windows 平台上的脚本
func (te *TaskExecutor) executeWindowsScript(taskID string, scriptPath string, logger *log.Logger) {
	fmt.Printf("正在 Windows 平台上执行任务 %s，脚本路径: %s\n", taskID, scriptPath)

	if strings.HasSuffix(scriptPath, ".bat") || strings.HasSuffix(scriptPath, ".cmd") {
		cmd := exec.Command("cmd", "/C", scriptPath)
		runScript(cmd, logger)
	} else if strings.HasSuffix(scriptPath, ".ps1") {
		cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
		runScript(cmd, logger)
	} else if strings.HasSuffix(scriptPath, ".py") {
		cmd := exec.Command("python", scriptPath)
		runScript(cmd, logger)
	} else if strings.HasSuffix(scriptPath, ".java") {
		cmd := exec.Command("java", scriptPath)
		runScript(cmd, logger)
	} else {
		logger.Println("未知的 Windows 脚本类型")
	}
}

// executeLinuxScript 执行 Linux 平台上的脚本
func (te *TaskExecutor) executeLinuxScript(taskID string, scriptPath string, logger *log.Logger) {
	fmt.Printf("正在 Linux 平台上执行任务 %s，脚本路径: %s\n", taskID, scriptPath)

	if strings.HasSuffix(scriptPath, ".sh") {
		cmd := exec.Command("/bin/sh", scriptPath)
		runScript(cmd, logger)
	} else if strings.HasSuffix(scriptPath, ".py") {
		cmd := exec.Command("python", scriptPath)
		runScript(cmd, logger)
	} else if strings.HasSuffix(scriptPath, ".java") {
		cmd := exec.Command("java", scriptPath)
		runScript(cmd, logger)
	} else {
		logger.Println("未知的 Linux 脚本类型")
	}
}

// runScript 运行脚本并捕获输出到日志文件
func runScript(cmd *exec.Cmd, logger *log.Logger) {
	// 创建管道捕获输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Printf("无法获取标准输出: %v\n", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Printf("无法获取标准错误输出: %v\n", err)
		return
	}

	// 启动命令执行
	if err := cmd.Start(); err != nil {
		logger.Printf("命令启动失败: %v\n", err)
		return
	}

	// 将标准输出和错误输出写入日志
	go io.Copy(logger.Writer(), stdout)
	go io.Copy(logger.Writer(), stderr)

	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		logger.Printf("命令执行失败: %v\n", err)
	} else {
		logger.Println("命令执行成功")
	}
}

func (te *TaskExecutor) getLogFilePath(taskID string) string {
	timestamp := time.Now().Format("20060102-150405")
	logDir := filepath.Join("logs", taskID)
	os.MkdirAll(logDir, os.ModePerm) // 确保日志目录存在
	return filepath.Join(logDir, fmt.Sprintf("%s_%s.log", taskID, timestamp))
}

// Stop 停止任务调度器
func (te *TaskExecutor) Stop() {
	te.CronScheduler.Stop() // 停止任务调度器
}
