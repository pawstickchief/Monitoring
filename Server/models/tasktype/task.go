package tasktype

import (
	"time"
)

type TaskRequest struct {
	TaskID     string `json:"task_id"`
	TaskName   string `json:"task_name"`
	ClientIP   string `json:"client_ip"`
	ScriptPath string `json:"script_path"`
	Action     string `json:"action"`
}

type TaskRecord struct {
	TaskID          string `json:"task_id" db:"task_id"`
	ClientIP        string `json:"client_ip" db:"client_ip"`
	ScriptPath      string `json:"script_path" db:"script_path"`
	Remarks         string `json:"remarks" db:"remarks"`
	CrondExpression string `json:"crond_expression" db:"crond_expression"`
	Status          string `json:"status" db:"status"` // 任务状态
	FileId          int64  `json:"file_id" db:"file_id"`
}

type ClientTaskLog struct {
	ClientIP       string    `json:"client_ip" db:"client_ip"`             // 客户端 IP
	TaskID         string    `json:"task_id" db:"task_id"`                 // 任务 ID，外键关联 task_records
	ExecutionTime  time.Time `json:"execution_time" db:"execution_time"`   // 任务执行时间
	Output         string    `json:"output" db:"output"`                   // 任务输出内容
	CompletionTime time.Time `json:"completion_time" db:"completion_time"` // 任务结束时间
	Remarks        string    `json:"remarks" db:"remarks"`                 // 备注
}
type TaskReceive struct {
	RequestId   string `json:"request_id"`
	TaskId      int    `json:"task_id"`
	TaskType    string `json:"task_type"`
	TaskBash    string `json:"task_bash"`
	TaskWorkDir string `json:"task_work_dir"`
}
type TaskRequestOption struct {
	Option      string      `json:"option" binding:"required"`
	Info        TaskReceive `json:"task_info"`
	FileId      int64       `json:"file_id"`
	Record      TaskRecord  `json:"task_record"`
	TaskControl TaskRequest `json:"task_control"`
}
