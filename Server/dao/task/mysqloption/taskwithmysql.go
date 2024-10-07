package mysqloption

import (
	"Server/models/tasktype"
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// InsertTaskRecord 插入一条新的任务记录到 task_records 表
func InsertTaskRecord(db *sqlx.DB, clientIP, scriptPath, remarks string, taskID string, crond string, fileId int64) (int64, error) {
	// 定义插入的 SQL 语句，使用命名参数
	query := `
		INSERT INTO task_records (client_ip, script_path, remarks, task_id,crond_expression,file_id)
		VALUES (:client_ip, :script_path, :remarks, :task_id,:crond_expression,:file_id)
	`

	// 创建一个 TaskRecord 实例，不包含 ID
	task := tasktype.TaskRecord{
		ClientIP:        clientIP,
		ScriptPath:      scriptPath,
		Remarks:         remarks,
		TaskID:          taskID,
		CrondExpression: crond,
		FileId:          fileId,
	}

	// 使用 NamedExec 进行命名参数的插入
	result, err := db.NamedExec(query, &task)
	if err != nil {
		zap.L().Error("Failed to insert task record", zap.Error(err))
		return 0, fmt.Errorf("failed to insert task record: %w", err)
	}

	// 获取插入的记录 ID
	id, err := result.LastInsertId()
	if err != nil {
		zap.L().Error("Failed to retrieve last insert ID", zap.Error(err))
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}

	zap.L().Info("Successfully inserted task record", zap.Int64("id", id))
	return id, nil
}

// InsertTaskLog 插入一条新的任务日志记录到 client_task_logs 表
func InsertTaskLog(db *sqlx.DB, taskLog *tasktype.ClientTaskLog) (int64, error) {
	// 定义插入的 SQL 语句，使用命名参数
	query := `
		INSERT INTO client_task_logs (client_ip, task_id, execution_time, output, completion_time, remarks)
		VALUES (:client_ip, :task_id, :execution_time, :output, :completion_time, :remarks)
	`

	// 使用 NamedExec 进行命名参数的插入
	result, err := db.NamedExec(query, taskLog)
	if err != nil {
		zap.L().Error("Failed to insert task log", zap.Error(err))
		return 0, fmt.Errorf("failed to insert task log: %w", err)
	}

	// 获取插入的记录 ID
	id, err := result.LastInsertId()
	if err != nil {
		zap.L().Error("Failed to retrieve last insert ID", zap.Error(err))
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}

	zap.L().Info("Successfully inserted task log", zap.Int64("id", id))
	return id, nil
}
