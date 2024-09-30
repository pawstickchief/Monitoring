package controller

import (
	"context"
	"fmt"
	"go-web-app/dao/etcd"
	"log"
	"time"
)

func GetTokenForClientFromController(clientIP string) (string, time.Time, error) {
	key := fmt.Sprintf("client_ip:%s", clientIP)

	// 输出查询的 key 信息
	log.Printf("Trying to get token for client IP: %s from etcd with key: %s", clientIP, key)

	// 在 etcd 中查找是否有该客户端的记录
	resp, err := etcd.GJobMgr.Kv.Get(context.Background(), key) // 调用 etcd 包中的 GJobMgr
	if err != nil {
		log.Printf("Error while getting key from etcd: %v", err)
		return "", time.Time{}, fmt.Errorf("failed to get key from etcd: %v", err)
	}

	// 如果有数据，打印 etcd 返回的数据
	if len(resp.Kvs) > 0 {
		etcdValue := string(resp.Kvs[0].Value)
		log.Printf("Found token in etcd: %s", etcdValue)

		tokenInfo, err := etcd.ParseTokenInfo(etcdValue)
		if err != nil {
			log.Printf("Failed to parse token info: %v", err)
			return "", time.Time{}, err
		}
		// 使用 ExpiresAtTime 字段调用 After 方法来比较时间
		if tokenInfo.ExpiresAtTime.After(time.Now()) {
			log.Printf("Token is valid, returning existing token for IP: %s", clientIP)
			return tokenInfo.Token, tokenInfo.ExpiresAtTime, nil
		} else {
			log.Printf("Token for IP %s has expired", clientIP)
		}
	} else {
		log.Printf("No token found for IP: %s in etcd", clientIP)
	}

	// 如果没有现有 token 或 token 已过期，则生成新 token
	newToken := etcd.GenerateToken(clientIP) // 调用 etcd 包中的 token 生成函数
	expiresAt := time.Now().Add(24 * time.Hour)

	// 将新 token 存入 etcd，并设置 24 小时的 TTL
	log.Printf("Storing new token for IP: %s with expiration time: %s", clientIP, expiresAt.Format(time.RFC3339))
	_, err = etcd.GJobMgr.Kv.Put(context.Background(), key, fmt.Sprintf("token:%s,expires_at:%s", newToken, expiresAt.Format(time.RFC3339)))
	if err != nil {
		log.Printf("Error while storing new token to etcd: %v", err)
		return "", time.Time{}, fmt.Errorf("failed to store new token in etcd: %v", err)
	}

	log.Printf("New token generated and stored for IP: %s, Token: %s, ExpiresAt: %s", clientIP, newToken, expiresAt.Format(time.RFC3339))

	return newToken, expiresAt, nil
}
