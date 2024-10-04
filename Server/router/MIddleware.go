package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func DBMiddleware(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 将 db 存储到 Context 中
		c.Set("db", db)
		c.Next()
	}
}
func ETCDMiddleware(cli *clientv3.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 将 db 存储到 Context 中
		c.Set("etcd", cli)
		c.Next()
	}
}
