package mysqloption

import (
	"Server/models/tasktype"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
)

func GetTaskRecordByTaskID(db *sqlx.DB, taskID string) (*tasktype.TaskRecord, error) {
	// 构建查询语句
	query := "SELECT script_path, task_id,crond_expression,remarks FROM task_records WHERE task_id = ?"

	// 执行查询
	row := db.QueryRow(query, taskID)

	// 解析查询结果
	var record tasktype.TaskRecord
	if err := row.Scan(&record.ScriptPath, &record.TaskID, &record.CrondExpression, &record.Remarks); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("没有找到任务 ID 为 %s 的记录", taskID)
		}
		return nil, fmt.Errorf("查询任务记录失败: %v", err)
	}

	return &record, nil
}

func GetTaskRecordsByTaskIDs(db *sqlx.DB, taskIDs []string) ([]tasktype.TaskRecord, error) {
	// 构造 IN 子句
	placeholders := make([]string, len(taskIDs))
	args := make([]interface{}, len(taskIDs))

	for i, id := range taskIDs {
		placeholders[i] = "?" // 占位符
		args[i] = id          // 参数
	}

	// 使用 IN 子句进行批量查询
	query := fmt.Sprintf("SELECT script_path, task_id, crond_expression, remarks,file_id FROM task_records WHERE task_id IN (%s)", strings.Join(placeholders, ","))

	// 执行查询
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("批量查询任务记录失败: %v", err)
	}
	defer rows.Close()

	// 解析查询结果
	var records []tasktype.TaskRecord
	for rows.Next() {
		var record tasktype.TaskRecord
		if err := rows.Scan(&record.ScriptPath, &record.TaskID, &record.CrondExpression, &record.Remarks, &record.FileId); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	// 检查 rows.Err() 以确保查询过程中没有发生错误
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}
