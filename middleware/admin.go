package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/zHElEARN/go-csust-planet/utils/response"
)

// AdminAuthMiddleware 后台身份验证中间件
func AdminAuthMiddleware(adminBearerToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.ResponseError(c, http.StatusUnauthorized, "未提供后台访问令牌")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && strings.EqualFold(parts[0], "Bearer")) {
			response.ResponseError(c, http.StatusUnauthorized, "后台访问令牌格式不正确")
			c.Abort()
			return
		}

		if parts[1] != adminBearerToken {
			response.ResponseError(c, http.StatusUnauthorized, "无效的后台访问令牌")
			c.Abort()
			return
		}

		c.Next()
	}
}
