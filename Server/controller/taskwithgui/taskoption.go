package taskwithgui

import (
	"Server/controller"
	"Server/dao/task/mysqloption"
	"Server/models/tasktype"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
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
	data, err := mysqloption.TaskOptionCore(p, db.(*sqlx.DB), cli.(*clientv3.Client))
	if err != nil {
		zap.L().Error("参数请求错误", zap.String("ParameterType", p.Option), zap.Error(err))
		controller.ResopnseError(c, controller.CodeUserNotExist)

		return
	}

	controller.ResopnseSystemDataSuccess(c, data)
}
