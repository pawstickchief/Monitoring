package task

import (
	"Server/dao/task/etcdoption"
	"context"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"strings"
)

// Manager 用于管理任务
type Manager struct {
	kv      clientv3.KV
	watcher clientv3.Watcher
}

// NewTaskManager 创建一个新的 Manager 实例
func NewTaskManager(cli *clientv3.Client) *Manager {
	return &Manager{
		kv:      clientv3.NewKV(cli),
		watcher: clientv3.NewWatcher(cli),
	}
}

// AddTasks 批量添加任务，提取公共逻辑，减少重复代码
func (tm *Manager) AddTasks(tasks map[string]string) error {
	if len(tasks) == 0 {
		return fmt.Errorf("任务列表为空")
	}

	err := etcdoption.BatchCreateIfNotExist(tm.kv, tasks)
	if err != nil {
		log.Printf("批量添加任务失败: %v", err)
		return err
	}

	log.Println("任务批量添加成功")
	return nil
}

// AddClientTask 为指定客户端添加任务
func (tm *Manager) AddClientTask(clientIP, taskID, taskStatus string) error {
	// 构建 etcd 中存储任务的键
	taskKey := fmt.Sprintf("/tasks/%s/%s", clientIP, taskID)

	// 启动一个 etcd 事务，确保操作的原子性
	txn := tm.kv.Txn(context.Background())

	// 添加任务
	txnResp, err := txn.If().Then(
		clientv3.OpPut(taskKey, taskStatus),
	).Commit()

	if err != nil {
		return fmt.Errorf("创建任务失败: %v", err)
	}

	if !txnResp.Succeeded {
		return fmt.Errorf("任务 %s 已经存在", taskID)
	}

	log.Printf("成功为客户端 %s 添加任务 %s", clientIP, taskID)
	return nil
}

// DeleteTasks 批量删除任务，添加更多日志信息
func (tm *Manager) DeleteTasks(taskKeys []string) error {
	if len(taskKeys) == 0 {
		return fmt.Errorf("要删除的任务列表为空")
	}

	err := etcdoption.BatchDeleteIfExist(tm.kv, taskKeys)
	if err != nil {
		log.Printf("批量删除任务失败: %v", err)
		return err
	}

	log.Println("任务批量删除成功")
	return nil
}

// WatchTasks 批量监听任务，优化日志信息并支持自定义回调
func (tm *Manager) WatchTasks(taskKeys []string, eventType mvccpb.Event_EventType, callback func(*mvccpb.Event)) {
	if len(taskKeys) == 0 {
		log.Println("没有任务需要监听")
		return
	}

	log.Println("任务监听启动")
	etcdoption.BatchWatchWithEvents(tm.watcher, taskKeys, eventType, callback)
}

// ManageTasks 批量创建、删除和监听任务，优化错误处理和日志记录
func (tm *Manager) ManageTasks(tasksToCreate map[string]string, tasksToDelete []string, tasksToWatch []string, eventType mvccpb.Event_EventType) {
	// 创建任务
	err := tm.AddTasks(tasksToCreate)
	if err != nil {
		log.Printf("任务创建失败: %v", err)
		return
	}

	// 删除任务
	err = tm.DeleteTasks(tasksToDelete)
	if err != nil {
		log.Printf("任务删除失败: %v", err)
		return
	}

	// 监听任务
	tm.WatchTasks(tasksToWatch, eventType, func(event *mvccpb.Event) {
		log.Printf("任务事件监听触发: %v", event)
	})
}

func (tm *Manager) GetTasksByClientIP(clientIP string, taskStatus string) ([]string, error) {
	// 使用客户端 IP 生成任务键的前缀
	keyPrefix := fmt.Sprintf("/tasks/%s", clientIP)

	// 从 etcd 获取所有与该客户端 IP 相关的任务
	taskInfos, err := tm.GetTasks(keyPrefix)
	if err != nil {
		log.Printf("从 etcd 获取任务失败: %v", err)
		return nil, err
	}

	// 如果没有找到任务，返回空列表
	if len(taskInfos) == 0 {
		log.Printf("没有为客户端 IP %s 找到任务", clientIP)
		return []string{}, nil // 返回空列表而不是错误
	}

	// 提取只包含匹配状态的任务 ID
	var matchingTaskIDs []string
	for _, taskInfo := range taskInfos {
		// 提取任务的值
		taskValue := string(taskInfo.Value)

		// 如果值为指定的任务状态（如 "inactive"），则提取任务 ID
		if taskValue == taskStatus {
			taskKey := string(taskInfo.Key)

			// 提取任务 ID（去掉路径前缀）
			taskIDParts := strings.Split(taskKey, "/")

			// 确保任务 ID 在路径的最后一部分，例如：/tasks/127.0.0.1/4810
			if len(taskIDParts) >= 3 {
				taskID := taskIDParts[len(taskIDParts)-1] // 提取最后一部分即为任务ID
				matchingTaskIDs = append(matchingTaskIDs, taskID)
			}
		}
	}

	// 返回匹配的任务 ID 列表
	return matchingTaskIDs, nil
}

// extractTaskIDs 从 taskInfos 中提取任务 ID，并返回字符串列表
func (tm *Manager) extractTaskIDs(taskInfos []*mvccpb.KeyValue) ([]string, error) {
	taskList := make([]string, 0)

	// 遍历从 etcd 获取的任务信息
	for _, taskInfo := range taskInfos {
		// 提取完整的任务键，例如 /tasks/192.168.1.100/8529
		taskKey := string(taskInfo.Key)

		// 提取任务 ID，也就是去掉路径前缀，保留最后一部分
		taskIDParts := strings.Split(taskKey, "/")
		if len(taskIDParts) < 3 {
			log.Printf("任务键格式无效: %s", taskKey)
			continue
		}

		// 倒数第一部分即为任务 ID
		taskID := taskIDParts[len(taskIDParts)-1]

		// 添加任务 ID 到列表
		taskList = append(taskList, taskID)
	}

	return taskList, nil
}

// GetTasks 从 etcd 获取指定前缀的任务列表
func (tm *Manager) GetTasks(prefix string) ([]*mvccpb.KeyValue, error) {
	// 使用 etcd 前缀查询任务
	resp, err := tm.kv.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("从 etcd 获取任务失败: %v", err)
	}

	// 返回等到的任务列表
	return resp.Kvs, nil
}

// assembleTasks 将任务列表转化为适当的格式
func (tm *Manager) assembleTasks(taskInfos []*mvccpb.KeyValue) ([]map[string]interface{}, error) {
	taskList := make([]map[string]interface{}, 0)

	for _, taskInfo := range taskInfos {
		taskID := string(taskInfo.Key)
		taskContent := string(taskInfo.Value)

		// 拼接任务信息
		task := map[string]interface{}{
			"task_id":      taskID,
			"task_content": taskContent,
		}

		// 将任务添加到任务列表中
		taskList = append(taskList, task)
	}

	return taskList, nil
}
