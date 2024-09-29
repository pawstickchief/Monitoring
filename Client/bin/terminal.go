package bin

import (
	"awesomeProject/mode"
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// PingWithCancel 执行带取消功能的 ping 命令
func PingWithCancel(ctx context.Context, address string, count int, outputChan chan<- string) error {
	// 构造 ping 命令，-n 参数用于 Windows，指定 ping 次数
	cmd := exec.CommandContext(ctx, "ping", address, "-n", fmt.Sprintf("%d", count))

	// 获取命令输出的管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	// 启动命令执行
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	// 逐行读取命令输出
	reader := bufio.NewReader(stdout)
	for {
		select {
		case <-ctx.Done():
			// 如果接收到取消信号，退出循环
			outputChan <- "Ping operation canceled by client"
			close(outputChan)
			return ctx.Err()
		default:
			// 继续读取命令的输出
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("error reading ping output: %v", err)
			}

			// 将 GBK 转换为 UTF-8
			utf8Line, err := mode.ConvertGBKToUTF8(line)
			if err != nil {
				return fmt.Errorf("failed to convert output to UTF-8: %v", err)
			}

			// 发送输出到 WebSocket
			outputChan <- strings.TrimSpace(utf8Line)
		}
	}

	// 等待命令结束
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command finished with error: %v", err)
	}

	close(outputChan) // 关闭通道，表示命令已完成
	return nil
}
