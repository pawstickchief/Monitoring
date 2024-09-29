package etcd

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"go-web-app/models"
	"go-web-app/models/clientsoket"
	"go-web-app/settings"
	clientv3 "go.etcd.io/etcd/client/v3"
	"io/ioutil"
	"time"
)

var (
	Client  *clientv3.Client
	Kv      clientv3.KV
	Lease   clientv3.Lease
	GJobMgr *models.JobMgr
)

func InitCrontab(cfg *settings.EtcdConfig) (err error) {
	// 加载 CA 证书
	caCert, err := ioutil.ReadFile(cfg.CaCert)
	if err != nil {
		fmt.Println("加载 CA 证书失败：", err)
		return
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		fmt.Println("解析 CA 证书失败")
		return
	}

	// 加载客户端证书和私钥
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		fmt.Println("加载客户端证书和私钥失败：", err)
		return
	}

	// 创建 TLS 配置
	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,              // 信任的 CA
		Certificates: []tls.Certificate{cert}, // 客户端证书
		ServerName:   cfg.ServerName,          // etcd 服务器的域名
	}

	// 配置 etcd 客户端
	config := clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: time.Duration(cfg.DialTimeout) * time.Millisecond,
		TLS:         tlsConfig,
		Username:    cfg.EtcdName,
		Password:    cfg.Password,
	}

	// 创建 etcd 客户端
	if Client, err = clientv3.New(config); err != nil {
		fmt.Println("连接 etcd 失败：", err)
		return err
	}

	// 获取 KV 和 Lease 的 API 子集
	Kv = clientv3.NewKV(Client)
	Lease = clientv3.NewLease(Client)

	// 赋值单例
	GJobMgr = &models.JobMgr{
		Clinet: Client,
		Kv:     Kv,
		Lease:  Lease,
	}

	return
}

// 解析 etcd 中存储的 token 信息
func ParseTokenInfo(data string) (tokenInfo clientsoket.TokenInfo) {
	// 假设 etcd 中存储的值是 "token:<token>,expires_at:<time>"
	fmt.Sscanf(data, "token:%s,expires_at:%s", &tokenInfo.Token, &tokenInfo.ExpiresAt)
	tokenInfo.ExpiresAtTime, _ = time.Parse(time.RFC3339, tokenInfo.ExpiresAt)
	return tokenInfo
}
func GenerateToken(clientIP string) string {
	// 获取当前时间
	timestamp := time.Now().Unix()

	// 将 IP 和时间戳结合在一起生成一个字符串
	data := fmt.Sprintf("%s:%d", clientIP, timestamp)

	// 使用 SHA256 生成哈希
	hash := sha256.New()
	hash.Write([]byte(data))
	token := hex.EncodeToString(hash.Sum(nil))

	// 返回生成的 token
	return token
}
