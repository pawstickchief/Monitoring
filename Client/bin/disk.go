package bin

import (
	"awesomeProject/datetype"
	"github.com/shirou/gopsutil/v3/disk"
	"go.uber.org/zap"
)

// GetDiskInfo 查询磁盘的信息，并返回结构化数据
func GetDiskInfo() ([]datetype.DiskData, error) {
	// 获取所有分区信息
	partitions, err := disk.Partitions(false)
	if err != nil {
		zap.L().Error("获取磁盘分区信息失败", zap.Error(err))
		return nil, err
	}

	var diskData []datetype.DiskData

	// 遍历每个分区并获取使用信息
	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			zap.L().Error("获取分区使用信息失败",
				zap.String("分区", partition.Mountpoint),
				zap.Error(err),
			)
			continue
		}

		// 创建 DiskData 实例并添加到返回切片中
		data := datetype.DiskData{
			Mountpoint:  partition.Mountpoint,
			Total:       float64(usage.Total),
			Used:        float64(usage.Used),
			Free:        float64(usage.Free),
			UsedPercent: usage.UsedPercent,
		}
		diskData = append(diskData, data)
	}

	// 返回磁盘数据
	return diskData, nil
}
