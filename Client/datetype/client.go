package datetype

type MemoryData struct {
	Total       float64 `json:"total"`
	Used        float64 `json:"used"`
	Free        float64 `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
}
type CPUData struct {
	PhysicalID string  `json:"physical_id"`
	Cores      int32   `json:"cores"`
	Mhz        float64 `json:"mhz"`
	Usage      float64 `json:"usage"`
}
type DiskData struct {
	Mountpoint  string  `json:"mountpoint"`
	Total       float64 `json:"total"`        // 总大小，单位字节
	Used        float64 `json:"used"`         // 已使用，单位字节
	Free        float64 `json:"free"`         // 可用，单位字节
	UsedPercent float64 `json:"used_percent"` // 使用率，百分比
}

type NetworkData struct {
	Name        string  `json:"name"`
	BytesSent   float64 `json:"bytesSent"`
	BytesRecv   float64 `json:"bytesRecv"`
	PacketsSent float64 `json:"packetsSent"`
	PacketsRecv float64 `json:"packetsRecv"`
}

type MonitorData struct {
	Memory  MemoryData    `json:"memory"`
	Network []NetworkData `json:"network"`
	CPU     []CPUData     `json:"cpu"`
	Disk    []DiskData    `json:"disk"`
}
