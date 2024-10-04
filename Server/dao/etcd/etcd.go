package etcd

import (
	"Server/models/clientsoket"
	"Server/models/tasktype"
	"Server/settings"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"io/ioutil"
	"strings"
	"time"
)

var (
	Client  *clientv3.Client
	Kv      clientv3.KV
	Lease   clientv3.Lease
	GJobMgr *tasktype.JobMgr
)

func InitCrontab(cfg *settings.EtcdConfig) (*clientv3.Client, error) {
	// 加载 CA 证书
	caCert, err := ioutil.ReadFile(cfg.CaCert)
	if err != nil {
		fmt.Println("加载 CA 证书失败：", err)
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		fmt.Println("解析 CA 证书失败")
		return nil, err
	}

	// 加载客户端证书和私钥
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		fmt.Println("加载客户端证书和私钥失败：", err)
		return nil, err
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
		return nil, err
	}
	// 初始化 Kv 和 Lease
	Kv = clientv3.NewKV(Client)
	Lease = clientv3.NewLease(Client)

	// 初始化全局 GJobMgr
	GJobMgr = &tasktype.JobMgr{
		Clinet: Client,
		Kv:     Kv,
		Lease:  Lease,
	}
	return Client, nil
}

// 解析 etcd 中存储的 token 信息

func ParseTokenInfo(data string) (tokenInfo clientsoket.TokenInfo, err error) {
	// 假设 etcd 中存储的值是 "token:<token>,expires_at:<time>"
	parts := strings.Split(data, ",")
	if len(parts) != 2 {
		return tokenInfo, fmt.Errorf("failed to parse token info: incorrect format")
	}

	// 分割 token 和 expires_at
	tokenPart := strings.Split(parts[0], "token:")
	expiresAtPart := strings.Split(parts[1], "expires_at:")
	if len(tokenPart) != 2 || len(expiresAtPart) != 2 {
		return tokenInfo, fmt.Errorf("failed to parse token info: incorrect format")
	}

	// 获取 token 和 expires_at 并解析过期时间
	tokenInfo.Token = strings.TrimSpace(tokenPart[1])
	tokenInfo.ExpiresAt = strings.TrimSpace(expiresAtPart[1])

	tokenInfo.ExpiresAtTime, err = time.Parse(time.RFC3339, tokenInfo.ExpiresAt)
	if err != nil {
		return tokenInfo, fmt.Errorf("failed to parse expiration time: %v", err)
	}

	return tokenInfo, nil
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
