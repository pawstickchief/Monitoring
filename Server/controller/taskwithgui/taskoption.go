package taskwithgui

import (
	"Server/common"
	"Server/controller"
	"Server/dao/task/mysqloption"
	"Server/models/tasktype"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"log"
)

func TaskManager(c *gin.Context) {
	db, exists := c.Get("db")
	if !exists {
		controller.ResopnseError(c, controller.CodeServerApiType)
		return
	}
	cli, exists := c.Get("etcd")
	if !exists {
		controller.ResopnseError(c, controller.CodeServerApiType)
		return
	}
	// 获取 WebSocket 管理器
	wsManager, exists := c.Get("wsManager") // 假设 WebSocket 管理器也存储在 gin.Context 中
	if !exists {
		controller.ResopnseError(c, controller.CodeServerApiType)
		return
	}
	manager, ok := wsManager.(*common.WebSocketManager) // 确保类型转换正确
	if !ok {
		log.Println("Invalid type for WebSocketManager in context")
		return
	}
	p := new(tasktype.TaskRequestOption)
	if err := c.ShouldBindJSON(&p); err != nil {
		//请求参数有误,直接返回响应
		var errs validator.ValidationErrors
		ok := errors.As(err, &errs)
		if !ok {
			controller.ResopnseError(c, controller.CodeServerApiType)
			return
		}
		controller.ResponseErrorwithMsg(c, controller.CodeServerApiType, controller.RemoveTopStruct(errs.Translate(controller.Trans)))
		return
	}
	data, err := mysqloption.TaskOptionCore(p, db.(*sqlx.DB), cli.(*clientv3.Client), manager.GetClients())
	if err != nil {
		zap.L().Error("参数请求错误", zap.String("ParameterType", p.Option), zap.Error(err))
		controller.ResopnseError(c, controller.CodeUserNotExist)

		return
	}

	controller.ResopnseSystemDataSuccess(c, data)
}
