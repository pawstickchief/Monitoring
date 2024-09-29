package controller

import (
	"context"
	"fmt"
	"go-web-app/dao/etcd"
	"time"
)

func GetTokenForClientFromController(clientIP string) (string, time.Time, error) {
	// 在 etcd 中查找是否有该客户端的记录
	key := fmt.Sprintf("client_ip:%s", clientIP)
	resp, err := etcd.GJobMgr.Kv.Get(context.Background(), key) // 调用 etcd 包中的 GJobMgr
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to get key from etcd: %v", err)
	}

	// 如果已有 token 且未过期，则返回现有的 token 和过期时间
	if len(resp.Kvs) > 0 {
		tokenInfo := etcd.ParseTokenInfo(string(resp.Kvs[0].Value))
		// 使用 ExpiresAtTime 字段调用 After 方法来比较时间
		if tokenInfo.ExpiresAtTime.After(time.Now()) {
			return tokenInfo.Token, tokenInfo.ExpiresAtTime, nil
		}
	}

	// 如果没有现有 token 或 token 已过期，则生成新 token
	newToken := etcd.GenerateToken(clientIP) // 调用 etcd 包中的 token 生成函数
	expiresAt := time.Now().Add(24 * time.Hour)

	// 将新 token 存入 etcd，并设置 24 小时的 TTL
	_, err = etcd.GJobMgr.Kv.Put(context.Background(), key, fmt.Sprintf("token:%s,expires_at:%s", newToken, expiresAt.Format(time.RFC3339)))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to store new token in etcd: %v", err)
	}

	return newToken, expiresAt, nil
}
