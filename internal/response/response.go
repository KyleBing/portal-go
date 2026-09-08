package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
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
