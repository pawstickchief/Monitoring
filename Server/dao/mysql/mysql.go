package mysql

import (
	"Server/settings"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

var db *sqlx.DB

// Init 初始化数据库连接，如果已经初始化则返回现有连接
func Init(cfg *settings.MySQLConfig) (*sqlx.DB, error) {
	// 如果 db 已经存在，直接返回
	if db != nil {
		return db, nil
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DbName)

	// 初始化 DB 对象并返回
	var err error
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		zap.L().Error("connect to db failed", zap.Error(err))
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	return db, nil
}

// GetDB 返回全局的数据库连接
func GetDB() *sqlx.DB {
	return db
}

// Close 关闭数据库连接
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
