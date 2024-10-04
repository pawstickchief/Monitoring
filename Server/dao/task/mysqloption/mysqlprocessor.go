package mysqloption

import (
	"Server/dao/task/etcdoption"
	"Server/models/tasktype"
	"Server/pkg/crondoption"
	"fmt"
	"github.com/jmoiron/sqlx"
	clientv3 "go.etcd.io/etcd/client/v3"
	"strconv"
)

func TaskOptionCore(p *tasktype.TaskRequestOption, db *sqlx.DB, cli *clientv3.Client) (interface{}, error) {
	taskid, _ := crondoption.GenerateFourDigitCodeFromIP(p.Record.ClientIP)
	// 构建要插入 etcd 的键值对
	keysValues := map[string]string{
		fmt.Sprintf("/tasks/%s/%s", p.Record.ClientIP, taskid): p.Record.Status,                 // 可以设置为任务状态
		fmt.Sprintf("/tasksfile/%s", taskid):                   strconv.FormatInt(p.FileId, 10), // 可以设置为任务状态
	}
	// 基于 Option 的值使用 switch 语句处理不同的逻辑
	switch p.Option {
	case "create":
		// 处理创建任务的逻辑
		etcdoption.BatchCreateIfNotExist(cli.KV, keysValues)
		record, err := InsertTaskRecord(db, p.Record.ClientIP, p.Record.ScriptPath, p.Record.Remarks, taskid, p.Record.CrondExpression)
		if err != nil {
			return nil, err
		}

		return record, nil

	case "update":
		// 处理更新任务的逻辑

		return nil, nil
	case "delete":
		// 处理删除任务的逻辑

		return nil, nil

	case "query":
		// 处理查询任务的逻辑

		return nil, nil

	default:
		// 如果 Option 是未知的，返回一个错误
		return nil, fmt.Errorf("未知的任务操作: %s", p.Option)
	}
}
