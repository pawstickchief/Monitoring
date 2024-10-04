package mode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// DownloadFile 函数用于下载多个文件并保存到指定的工作目录
func DownloadFile(fileIDs []string, address string, workdir string) error {

	for _, fileIDStr := range fileIDs {
		fileID, err := strconv.Atoi(fileIDStr)
		if err != nil {
			return fmt.Errorf("转换 fileID '%s' 为 int 失败: %v", fileIDStr, err)
		}
		// 构建下载 URL
		downloadURL := fmt.Sprintf("%s", address) // 假设下载地址为 /download

		// 创建请求体
		requestBody, err := json.Marshal(map[string]int{
			"fileid": fileID,
		})
		if err != nil {
			return fmt.Errorf("创建请求体失败: %v", err)
		}

		// 发送 POST 请求进行下载
		resp, err := http.Post(downloadURL, "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			return fmt.Errorf("下载任务 %s 失败: %v", fileID, err)
		}
		defer resp.Body.Close()

		// 从响应头获取文件名
		contentDisposition := resp.Header.Get("Content-Disposition")
		var fileName string
		if contentDisposition != "" {
			_, params, err := mime.ParseMediaType(contentDisposition)
			if err == nil {
				fileName = params["filename"]
			}
		}

		// 如果无法从响应头获取文件名，使用 fileID 作为默认文件名
		if fileName == "" {
			fileName = fmt.Sprintf("%d_taskfile", fileID) // 没有扩展名的默认文件名
		}

		// 构建保存文件的路径
		filePath := filepath.Join(workdir, fileName)

		// 创建本地文件
		out, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("创建文件失败: %v", err)
		}
		defer out.Close()

		// 将下载的内容写入文件
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return fmt.Errorf("文件写入失败: %v", err)
		}

		fmt.Printf("文件 %s 已成功下载到 %s\n", fileName, filePath)
	}

	return nil
}

// UploadFiles 函数用于将多个文件上传到指定的服务器地址
func UploadFiles(filePaths []string, address string) error {

	for _, filePath := range filePaths {
		// 打开文件
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("无法打开文件 '%s': %v", filePath, err)
		}
		defer file.Close()

		// 创建一个缓冲区存储 multipart/form-data 请求
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// 添加文件字段到 multipart 表单
		part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
		if err != nil {
			return fmt.Errorf("创建 multipart 表单失败: %v", err)
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return fmt.Errorf("写入文件到 multipart 表单失败: %v", err)
		}

		// 关闭 multipart writer，以完成请求体的构建
		err = writer.Close()
		if err != nil {
			return fmt.Errorf("关闭 multipart writer 失败: %v", err)
		}

		// 发送 POST 请求进行文件上传
		req, err := http.NewRequest("POST", address, body)
		if err != nil {
			return fmt.Errorf("创建 HTTP 请求失败: %v", err)
		}

		// 设置 Content-Type 为 multipart/form-data
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("文件上传任务失败: %v", err)
		}
		defer resp.Body.Close()

		// 读取响应
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("读取服务器响应失败: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("服务器响应错误: %s, 响应内容: %s", resp.Status, string(respBody))
		}

		fmt.Printf("文件 %s 已成功上传到服务器\n", filepath.Base(filePath))
	}

	return nil
}
