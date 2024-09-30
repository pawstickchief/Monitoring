package task

import (
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"log"
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

	err := BatchCreateIfNotExist(tm.kv, tasks)
	if err != nil {
		log.Printf("批量添加任务失败: %v", err)
		return err
	}

	log.Println("任务批量添加成功")
	return nil
}

// DeleteTasks 批量删除任务，添加更多日志信息
func (tm *Manager) DeleteTasks(taskKeys []string) error {
	if len(taskKeys) == 0 {
		return fmt.Errorf("要删除的任务列表为空")
	}

	err := BatchDeleteIfExist(tm.kv, taskKeys)
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
	BatchWatchWithEvents(tm.watcher, taskKeys, eventType, callback)
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
