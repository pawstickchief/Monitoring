package task

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// 批量新增，不覆盖现有键

func BatchCreateIfNotExist(kv clientv3.KV, keysValues map[string]string) error {
	for key, value := range keysValues {
		// 检查键是否存在
		resp, err := kv.Get(context.Background(), key)
		if err != nil {
			return err
		}
		// 如果键不存在，则创建
		if resp.Count == 0 {
			_, err := kv.Put(context.Background(), key, value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
