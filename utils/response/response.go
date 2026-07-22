package response

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// ResponseError 统一的错误响应格式
func ResponseError(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponse{Error: message})
}

func HandleContextError(c *gin.Context, err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		ResponseError(c, http.StatusGatewayTimeout, "请求处理超时")
		c.Abort()
		return true
	}
	if errors.Is(err, context.Canceled) {
		c.Abort()
		return true
	}
	return false
}
