package common

import (
	"github.com/gin-gonic/gin"
	//"net/http"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, statusCode int, message string, errCode string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
		Error:   errCode,
	})
}
