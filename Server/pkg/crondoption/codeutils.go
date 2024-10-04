package crondoption

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"
)

// GenerateFourDigitCodeFromIP 根据客户端 IP 地址生成 4 位数字
func GenerateFourDigitCodeFromIP(clientIP string) (string, error) {
	// 解析 IP 地址
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return "", fmt.Errorf("无效的 IP 地址: %s", clientIP)
	}

	// 获取当前时间的纳秒数作为随机因子
	timeNano := time.Now().UnixNano()

	// 将 IP 和时间戳组合生成哈希值
	data := fmt.Sprintf("%s:%d", clientIP, timeNano)
	hash := sha256.New()
	hash.Write([]byte(data))
	hashed := hash.Sum(nil)

	// 从哈希值中提取前两个字节（16 位），并将其转换为整数
	number := binary.BigEndian.Uint16(hashed[:2])

	// 对 10000 取模，确保结果是 4 位数字
	fourDigitCode := int(number % 10000)

	// 将 int 转换为 string 并返回
	taskID := strconv.Itoa(fourDigitCode)
	return taskID, nil
}
