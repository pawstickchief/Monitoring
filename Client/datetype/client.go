package datetype

import "time"

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
type TaskStatus struct {
	RequestId  string `json:"request_id"`
	TaskType   string `json:"task_type"`
	TaskId     string `json:"task_id"`
	TaskStatus string `json:"task_status"`
}
type TaskLogRequest struct {
	RequestId     string `json:"request_id"`
	TaskType      string `json:"task_type"`
	ClientTaskLog `json:"client_task_log"`
}
type TaskRequest struct {
	RequestId string   `json:"request_id"`
	TaskId    []string `json:"task_id"`
	TaskType  string   `json:"task_type"`
}
type TaskReceive struct {
	RequestId   string `json:"request_id"`
	TaskId      string `json:"task_id"`
	TaskType    string `json:"task_type"`
	TaskBash    string `json:"task_bash"`
	TaskWorkDir string `json:"task_work_dir"`
}
type ClientTaskLog struct {
	LogID          int       `json:"log_id" db:"log_id"`                   // 日志 ID
	ClientIP       string    `json:"client_ip" db:"client_ip"`             // 客户端 IP
	TaskID         string    `json:"task_id" db:"task_id"`                 // 任务 ID，外键关联 task_records
	ExecutionTime  time.Time `json:"execution_time" db:"execution_time"`   // 任务执行时间
	Output         string    `json:"output" db:"output"`                   // 任务输出内容
	CompletionTime time.Time `json:"completion_time" db:"completion_time"` // 任务结束时间
	Remarks        string    `json:"remarks" db:"remarks"`                 // 备注
}
