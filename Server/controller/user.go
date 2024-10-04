package controller

import (
	"Server/dao/mysql"
	"Server/models"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// LoginUserVerif 查询交换机端口
// @Summary 查询交换机端口
// @Description 查询交换机端口
// @Accept  json
// @Produce  json
// @Param data body models.SelectSwitchMac true "查询交换机的参数"
// @Success 200 {object} models.ClientSwitchInfo "成功"
// @Failure 500 {object} models.ErrorResponse "内部错误"
// @Router /Login [post]

func LoginUserVerif(c *gin.Context) {
	p := new(models.LoginUserinfo)
	if err := c.ShouldBindJSON(&p); err != nil {
		//请求参数有误,直接返回响应
		var errs validator.ValidationErrors
		ok := errors.As(err, &errs)
		if !ok {
			ResopnseError(c, CodeServerApiType)
			return
		}
		ResponseErrorwithMsg(c, CodeServerApiType, RemoveTopStruct(errs.Translate(Trans)))
		return
	}
	var userinfo models.User
	userinfo.Name = p.UserName
	userinfo.UserId = p.UserCode
	err := mysql.LoginCode(&userinfo)
	if err != nil {
		zap.L().Error("用户信息核对失败", zap.String("ParameterType", p.UserName), zap.Error(err))
		ResopnseError(c, CodeUserNotExist)

		return
	}
	//3.返回响应

	ResopnseSystemDataSuccess(c, "用户信息核对正确")
}
