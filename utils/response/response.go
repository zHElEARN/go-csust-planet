package response

import (
	"github.com/gin-gonic/gin"
)

// ResponseError 统一的错误响应格式
func ResponseError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"error": message,
	})
}

// ResponseSuccess 统一的成功响应格式
func ResponseSuccess(c *gin.Context, message string, payload ...any) {
	resp := gin.H{
		"message": message,
	}
	if len(payload) > 0 {
		resp["payload"] = payload[0]
	}
	c.JSON(200, resp)
}
