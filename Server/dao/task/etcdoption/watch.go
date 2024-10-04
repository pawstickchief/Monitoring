package etcdoption

import (
	"context"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// 批量监听指定的事件类型 (PUT/DELETE)，并执行自定义回调函数

func BatchWatchWithEvents(watcher clientv3.Watcher, keys []string, eventType mvccpb.Event_EventType, callback func(ev *mvccpb.Event)) {
	for _, key := range keys {
		go func(k string) {
			watchChan := watcher.Watch(context.Background(), k)

			for watchResp := range watchChan {
				for _, ev := range watchResp.Events {
					// 只处理指定类型的事件
					if ev.Type == eventType {
						// 调用回调函数处理事件
						callback((*mvccpb.Event)(ev))
					}
				}
			}
		}(key)
	}
}
