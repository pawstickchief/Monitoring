package crond

type TaskRequest struct {
	TaskID   string `json:"task_id"`
	TaskName string `json:"task_name"`
	Payload  string `json:"payload"`
}
