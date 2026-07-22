package response

import "github.com/gin-gonic/gin"

type ErrorResponse struct {
	Error string `json:"error"`
}

// ResponseError 统一的错误响应格式
func ResponseError(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponse{Error: message})
}
