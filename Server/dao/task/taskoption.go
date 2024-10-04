package task

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"strings"
)

// CreateTaskForClient 创建一个任务并将其关联到客户端 IP
func (tm *Manager) CreateTaskForClient(clientIP, taskID string) error {
	// 生成任务状态键和客户端任务对应关系键
	taskStatusKey := fmt.Sprintf("/tasks/%s/status", taskID)
	clientTaskKey := fmt.Sprintf("/clients/%s/tasks", clientIP)

	// 获取当前客户端已经关联的任务（如果有）
	resp, err := tm.kv.Get(context.Background(), clientTaskKey)
	if err != nil {
		return fmt.Errorf("获取客户端任务失败: %v", err)
	}

	var tasks []string
	if len(resp.Kvs) > 0 {
		// 如果客户端已经有任务，将新任务追加到已有任务列表中
		tasks = strings.Split(string(resp.Kvs[0].Value), ",")
	}
	tasks = append(tasks, taskID)

	// 开始事务
	txn := tm.kv.Txn(context.Background())

	// 存储客户端与任务的关联关系，并创建任务状态
	txnResp, err := txn.If().
		Then(
			clientv3.OpPut(taskStatusKey, "pending"),                // 初始化任务状态为 pending
			clientv3.OpPut(clientTaskKey, strings.Join(tasks, ",")), // 更新客户端关联的任务列表
		).Commit()

	if err != nil {
		return fmt.Errorf("创建任务失败: %v", err)
	}

	if !txnResp.Succeeded {
		return fmt.Errorf("任务创建失败，事务未成功")
	}

	log.Printf("任务 %s 已成功创建并关联到客户端 %s", taskID, clientIP)
	return nil
}

// DeleteTask 删除任务状态
func (tm *Manager) DeleteTask(taskID string) error {
	// 生成任务状态的键
	taskStatusKey := fmt.Sprintf("/tasks/%s/status", taskID)

	// 开始事务
	txn := tm.kv.Txn(context.Background())

	// 删除任务状态
	_, err := txn.Then(
		clientv3.OpDelete(taskStatusKey),
	).Commit()

	if err != nil {
		return fmt.Errorf("删除任务失败: %v", err)
	}

	log.Printf("任务 %s 的状态已成功删除", taskID)
	return nil
}

// UpdateTaskStatus 更新任务状态
func (tm *Manager) UpdateTaskStatus(taskID string, newStatus string) error {
	// 生成任务状态的键
	taskStatusKey := fmt.Sprintf("/tasks/%s/status", taskID)

	// 检查任务状态是否存在
	resp, err := tm.kv.Get(context.Background(), taskStatusKey)
	if err != nil {
		return fmt.Errorf("检查任务状态失败: %v", err)
	}

	if len(resp.Kvs) == 0 {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	// 更新任务状态
	_, err = tm.kv.Put(context.Background(), taskStatusKey, newStatus)
	if err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	log.Printf("任务 %s 的状态已更新为 %s", taskID, newStatus)
	return nil
}

// GetTaskStatus 查询多个任务的状态
func (tm *Manager) GetTaskStatus(taskIDs []string) (map[string]string, error) {
	taskStatuses := make(map[string]string)

	for _, taskID := range taskIDs {
		// 生成任务状态键
		taskStatusKey := fmt.Sprintf("/tasks/%s/status", taskID)

		// 查询任务状态
		resp, err := tm.kv.Get(context.Background(), taskStatusKey)
		if err != nil {
			return nil, fmt.Errorf("查询任务状态失败: %v", err)
		}

		if len(resp.Kvs) == 0 {
			taskStatuses[taskID] = "任务不存在"
		} else {
			taskStatuses[taskID] = string(resp.Kvs[0].Value)
		}
	}

	log.Printf("任务状态查询结果: %v", taskStatuses)
	return taskStatuses, nil
}
