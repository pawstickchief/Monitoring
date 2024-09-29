
## Etcd用户初始化
```
export ETCDCTL_API=3
export ETCDCTL_ENDPOINTS="https://servername:2379"
export ETCDCTL_CACERT="/etc/etcd/ssl/chain.pem"
export ETCDCTL_CERT="/etc/etcd/ssl/fullchain.pem"
export ETCDCTL_KEY="/etc/etcd/ssl/privkey.pem"
export ETCDCTL_USER="root:your_password"  # 如果启用了认证
/usr/local/etcd/etcd --config-file /etc/etcd/etcd.yml 2>&1 &
./etcdctl user add admin
./etcdctl role add root
./etcdctl role grant-permission root --prefix=true readwrite /
./etcdctl user grant-role admin root
./etcdctl user add root
```

## 证书域名申请

```
yum install certbot
certbot certonly --standalone -d servername
```