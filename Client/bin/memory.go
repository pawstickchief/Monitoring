package bin

import (
	"Client/datetype"
	"github.com/shirou/gopsutil/v3/mem"
)

func GetMemoryData() (datetype.MemoryData, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return datetype.MemoryData{}, err
	}

	memData := datetype.MemoryData{
		Total:       float64(v.Total),
		Used:        float64(v.Used),
		Free:        float64(v.Free),
		UsedPercent: v.UsedPercent,
	}

	return memData, nil
}
