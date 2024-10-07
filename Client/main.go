package main

import (
	"Client/logger"
	"Client/route"
	"Client/setting"
	"Client/taskexce"
	"Client/taskmanager"
	"Client/ws"
	"context"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var appConfigpath string
	flag.StringVar(&appConfigpath, "c", "", "Configuration file path")
	flag.Parse()

	// 1. 加载配置文件
	if err := setting.Init(appConfigpath); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}

	// 2. 初始化日志
	if err := logger.Init(setting.Conf.LogConfig); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	defer func(l *zap.Logger) {
		err := l.Sync()
		if err != nil {
			zap.L().Error("L.Sync failed...")
		}
	}(zap.L())
	zap.L().Debug("logger init success...")

	// 3. 注册路由
	r := route.Setup()

	// 4. 启动 HTTP 服务（优雅关闭）
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", setting.Conf.Port),
		Handler: r,
	}
	// 启动 HTTP 服务
	go startServer(srv)

	// 5. 初始化任务管理器和任务执行器
	initTaskManagerAndExecutor()

	// 6. 等待中断信号来优雅地关闭服务器
	gracefulShutdown(srv)
}

// 启动 HTTP 服务
func startServer(srv *http.Server) {
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
}

// 优雅关闭服务器
func gracefulShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1) // 创建一个接收信号的通道
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // 等待信号
	zap.L().Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server Shutdown: ", zap.Error(err))
	}
	zap.L().Info("Server exiting")
}

func initTaskManagerAndExecutor() {
	// 创建任务管理器实例
	tm := &taskmanager.TaskManager{
		TaskList: make(map[string]taskmanager.Task), // 初始化任务列表的 map
	}

	// 连接 WebSocket 并请求 Token
	client, err := ws.ConnectWebSocketAndRequestToken(setting.Conf.ServerConfig.Ip, setting.Conf.ClientIp, setting.Conf.ServerConfig.Port)
	if err != nil {
		log.Fatalf("Failed to connect WebSocket: %v", err)
	}

	// 初始化 WebSocketManager
	wsManager := ws.GetWebSocketManager(tm)
	wsManager.Client = client // 确保 WebSocketClient 被正确赋值

	// 启动心跳检测协程
	go ws.Heartbeat(client)

	// 调用 InitTask 初始化任务列表
	tm.WSManager = wsManager // 确保 WSManager 已赋值
	tm.InitTask()            // 从服务器获取任务并初始化任务列表
	// 初始化任务执行器，并将 WebSocketManager 和 TaskManager 分开传递
	executor := taskexce.InitExecutor(tm, wsManager)
	executor.Start()

	go func() {
		ws.ListenToServerAndManageTasks(client)
	}()
}
