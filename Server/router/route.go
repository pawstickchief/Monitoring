package router

import (
	"Server/common"
	"Server/controller"
	"Server/controller/taskwithgui"
	"Server/dao/mysql"
	"Server/logger"
	"Server/middlewares"
	"Server/models"
	"Server/ws"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net/http"
)

func Setup(mode, ClientUrl string, size int64, savedir string, db *sqlx.DB, cli *clientv3.Client, wsManager *common.WebSocketManager) *gin.Engine {
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(WebSocketManagerMiddleware(wsManager))
	r.Use(DBMiddleware(db))
	r.Use(ETCDMiddleware(cli))
	r.Use(middlewares.Cors(ClientUrl))
	r.Use(logger.GinLogger(), logger.GinRecovery(true))
	r.MaxMultipartMemory = size << 20
	//注册业务路由
	// 注册Swagger路由
	// 映射 Swagger UI 相关静态文件
	r.Static("/swagger-ui", "./docs/swagger-ui")
	r.StaticFile("/swagger.json", "./docs/swagger.json")

	url := ginSwagger.URL("/swagger.json")
	r.POST("/TaskManager", taskwithgui.TaskManager)
	r.GET("/wsclient", ws.WebsocketHandler)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	r.POST("/login", controller.LoginUserVerif)
	r.POST("/download", controller.DownloadHandler)
	r.POST("/control", ws.ControlClientTask)
	r.POST("/upload", func(ctx *gin.Context) {
		forms, err := ctx.MultipartForm()
		if err != nil {
			fmt.Println("error", err)
		}
		files := forms.File["file"]
		var fileInfos []map[string]interface{}
		for _, v := range files {
			filelog := &models.Filelog{
				FileName: v.Filename,
				FileSize: v.Size,
				FileDir:  savedir + v.Filename,
			}
			if err := ctx.SaveUploadedFile(v, fmt.Sprintf("%s%s", savedir, v.Filename)); err != nil {
				ctx.String(http.StatusBadRequest, fmt.Sprintf("upload err %s", err.Error()))
			}
			fileid, _ := mysql.FileLogAdd(filelog)
			// 添加文件信息到返回列表
			fileInfos = append(fileInfos, map[string]interface{}{
				"file_name": v.Filename,
				"file_id":   fileid,
			})
		}
		// 返回上传成功的文件信息
		ctx.JSON(http.StatusOK, gin.H{
			"status": "success",
			"files":  fileInfos,
		})
	})
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})

	return r
}
