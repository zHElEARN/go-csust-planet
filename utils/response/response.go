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
