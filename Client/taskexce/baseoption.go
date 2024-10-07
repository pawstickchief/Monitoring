package taskexce

import (
	"Client/datetype"
	"Client/setting"
	"Client/ws"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

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

	// 更新任务状态为 "running"
	_, err = ws.UpdateTaskStatus(te.wsManager, taskID, "running")
	if err != nil {
		logger.Printf("更新任务 %s 的状态为 'running' 失败: %v\n", taskID, err)
		return
	}

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

	// 更新任务状态为 "completed"
	_, err = ws.UpdateTaskStatus(te.wsManager, taskID, "completed")
	if err != nil {
		logger.Printf("更新任务 %s 的状态为 'completed' 失败: %v\n", taskID, err)
		return
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
	_, err = ws.TaskLogPut(te.wsManager, taskLog)
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
