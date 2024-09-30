package main

import (
	"context"
	"flag"
	"fmt"
	"go-web-app/controller"
	"go-web-app/dao/etcd"
	"go-web-app/dao/mysql"
	"go-web-app/dao/task"
	"go-web-app/logger"
	"go-web-app/router"
	"go-web-app/settings"
	"go-web-app/ws"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Go Web开发比较通用的脚手架模板
// @title 交换机查询设备端口
// @version 1.0
// @description 通过mac地址查询对应设备所处交换机端口和vlan信息
// @host 192.168.8.84:8081
// @BasePath /
func main() {
	var appConfigPath string
	flag.StringVar(&appConfigPath, "c", "", "Configuration file path")
	flag.Parse()

	// 初始化配置
	if err := initConfig(appConfigPath); err != nil {
		fmt.Println(err)
		return
	}

	// 初始化日志
	if err := initLogger(); err != nil {
		fmt.Println(err)
		return
	}
	defer zap.L().Sync()

	// 初始化 etcd 和 mysql
	cli, err := initEtcd()
	if err != nil {
		zap.L().Error("init etcd failed", zap.Error(err))
		return
	}
	// 初始化 TaskManager
	task.NewTaskManager(cli)

	// 初始化处理器
	ws.InitHandlers()

	if err := initDatabase(); err != nil {
		zap.L().Error("init database failed", zap.Error(err))
		return
	}
	defer mysql.Close()

	// 初始化 Gin 的翻译器
	if err := controller.InitTrans("zh"); err != nil {
		zap.L().Error("init validator failed", zap.Error(err))
		return
	}

	// 注册路由
	r := router.Setup(settings.Conf.Mode, settings.Conf.ClientUrl, settings.Conf.Filemaxsize, settings.Conf.Savedir)

	// 启动 HTTP 服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", settings.Conf.Port),
		Handler: r,
	}
	startServer(srv)

	// 优雅关机处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	gracefulShutdown(srv, quit)
}
func startServer(srv *http.Server) {
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("listen: %v", zap.Error(err))
		}
	}()
}

func gracefulShutdown(srv *http.Server, quit chan os.Signal) {
	<-quit
	zap.L().Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server Shutdown: ", zap.Error(err))
	}

	zap.L().Info("Server exiting")
}
func initConfig(appConfigPath string) error {
	if err := settings.Init(appConfigPath); err != nil {
		return fmt.Errorf("init settings failed: %v", err)
	}
	return nil
}
func initLogger() error {
	if err := logger.Init(settings.Conf.LogConfig, settings.Conf.Mode); err != nil {
		return fmt.Errorf("init logger failed: %v", err)
	}
	return nil
}
func initDatabase() error {
	if err := mysql.Init(settings.Conf.MySQLConfig); err != nil {
		return fmt.Errorf("init mysql failed: %v", err)
	}
	return nil
}

func initEtcd() (*clientv3.Client, error) {
	client, err := etcd.InitCrontab(settings.Conf.EtcdConfig)
	if err != nil {
		return nil, fmt.Errorf("init etcd failed: %v", err)
	}
	return client, nil
}
