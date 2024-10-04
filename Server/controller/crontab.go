package controller

import (
	"Server/dao/mysql"
	"Server/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"os"
)

func DownloadHandler(c *gin.Context) {
	p := new(models.Filelog)
	if err := c.ShouldBindJSON(&p); err != nil {
		// 请求参数有误, 直接返回响应
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResopnseError(c, CodeServerApiType)
			return
		}
		ResponseErrorwithMsg(c, CodeServerApiType, RemoveTopStruct(errs.Translate(Trans)))
		return
	}

	// 获取文件名和文件路径
	attchmentName := mysql.FileName(p.FileId)
	attchmentDir := mysql.FileDir(p.FileId)

	// 尝试打开文件，检查是否存在
	file, err := os.Open(attchmentDir)
	if err != nil {
		fmt.Printf("文件获取失败: %s\n", attchmentDir)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found or could not be opened.",
		})
		return
	}
	defer file.Close()

	// 设置响应头
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", attchmentName))
	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", getFileSize(file))) // 设置文件大小

	// 发送文件
	c.File(attchmentDir)
}

// 获取文件大小的辅助函数
func getFileSize(file *os.File) int64 {
	info, err := file.Stat()
	if err != nil {
		return 0
	}
	return info.Size()
}
