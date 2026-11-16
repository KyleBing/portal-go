package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	// CodeUnauthorized token 缺失、无效或过期
	CodeUnauthorized = "unauthorized"
)

type Body struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Code    string      `json:"code,omitempty"`
}

func Success(c *gin.Context, data interface{}, message string) {
	if message == "" {
		message = "请求成功"
	}
	c.JSON(http.StatusOK, Body{Success: true, Message: message, Data: data})
}

func Error(c *gin.Context, data interface{}, message string) {
	if message == "" {
		message = "请求失败"
	}
	c.JSON(http.StatusOK, Body{Success: false, Message: message, Data: data})
}

// Unauthorized token 验证失败：HTTP 401 + code=unauthorized
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "身份验证失败：token 无效或已过期"
	}
	c.JSON(http.StatusUnauthorized, Body{
		Success: false,
		Message: message,
		Data:    nil,
		Code:    CodeUnauthorized,
	})
}
