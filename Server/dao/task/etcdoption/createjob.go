package etcdoption

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
)

// 批量新增，不覆盖现有键

func BatchCreateIfNotExist(kv clientv3.KV, keysValues map[string]string) error {
	// 创建一个 etcd 事务，确保操作的原子性
	txn := kv.Txn(context.Background())

	// 构建 etcd 操作的条件和操作
	ops := make([]clientv3.Op, 0)

	// 循环处理每一个键值对
	for key, value := range keysValues {
		// 检查键是否已经存在
		getResp, err := kv.Get(context.Background(), key)
		if err != nil {
			return fmt.Errorf("获取键 %s 失败: %v", key, err)
		}

		// 如果键不存在，则将其加入事务操作中
		if getResp.Count == 0 {
			ops = append(ops, clientv3.OpPut(key, value))
		}
	}

	// 如果没有要创建的键，则直接返回
	if len(ops) == 0 {
		log.Println("没有需要创建的键")
		return nil
	}

	// 在事务中添加操作
	txn = txn.Then(ops...)

	// 提交事务
	txnResp, err := txn.Commit()
	if err != nil {
		return fmt.Errorf("批量创建任务失败: %v", err)
	}

	if !txnResp.Succeeded {
		return fmt.Errorf("批量创建任务时失败")
	}

	log.Println("批量任务创建成功")
	return nil
}
