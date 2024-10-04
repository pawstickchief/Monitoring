package bin

import (
	"Client/datetype" // 修改为你的项目路径
	"github.com/shirou/gopsutil/v3/cpu"
	"go.uber.org/zap"
	"time"
)

// GetCPUInfo 获取 CPU 信息并返回数据结构
func GetCPUInfo() ([]datetype.CPUData, error) {
	// 获取 CPU 信息
	info, err := cpu.Info()
	if err != nil {
		zap.L().Error("获取 CPU 信息失败", zap.Error(err))
		return nil, err
	}

	// 获取 CPU 使用率信息
	usage, err := cpu.Percent(1*time.Second, false)
	if err != nil {
		zap.L().Error("获取 CPU 使用率失败", zap.Error(err))
		return nil, err
	}

	// 定义返回的 CPU 数据
	var cpuData []datetype.CPUData

	// 遍历 CPU 信息
	for i, cpuInfo := range info {
		data := datetype.CPUData{
			PhysicalID: cpuInfo.PhysicalID,
			Cores:      cpuInfo.Cores,
			Mhz:        cpuInfo.Mhz,
			Usage:      usage[i], // 使用 CPU 使用率
		}
		cpuData = append(cpuData, data)
	}

	// 返回 CPU 信息
	return cpuData, nil
}
