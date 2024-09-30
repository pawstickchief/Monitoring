package task

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// 批量删除，只删除存在的键

func BatchDeleteIfExist(kv clientv3.KV, keys []string) error {
	for _, key := range keys {
		// 检查键是否存在
		resp, err := kv.Get(context.Background(), key)
		if err != nil {
			return err
		}
		// 如果键存在，则删除
		if resp.Count > 0 {
			_, err := kv.Delete(context.Background(), key)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
