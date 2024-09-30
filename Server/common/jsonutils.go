package common

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

// SendJSONResponse 发送 JSON 格式的数据到 WebSocket 连接
func SendJSONResponse(conn *websocket.Conn, message map[string]interface{}) error {
	// 序列化消息为 JSON 格式
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return err
	}

	// 发送 JSON 消息到 WebSocket 连接
	err = conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		log.Printf("Error sending message: %v", err)
		return err
	}

	return nil
}

// ConvertToInterfaceMap 将 map[string]string 转换为 map[string]interface{}
func ConvertToInterfaceMap(input map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range input {
		result[key] = value
	}
	return result
}

// 将 map[string]interface{} 转换为 map[string]string
func convertInterfaceToStringMap(input map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for key, value := range input {
		strValue, ok := value.(string)
		if ok {
			result[key] = strValue
		} else {
			result[key] = fmt.Sprintf("%v", value) // 如果不是字符串，转换为字符串
		}
	}
	return result
}
