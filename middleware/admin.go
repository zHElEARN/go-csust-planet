package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zHElEARN/go-csust-planet/utils/response"
)

// AdminAuthMiddleware 后台身份验证中间件
func AdminAuthMiddleware(adminBearerToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := parseBearerToken(c.GetHeader("Authorization"))
		if errors.Is(err, errAuthorizationHeaderMissing) {
			response.ResponseError(c, http.StatusUnauthorized, "未提供后台访问令牌")
			c.Abort()
			return
		}
		if errors.Is(err, errAuthorizationHeaderInvalid) {
			response.ResponseError(c, http.StatusUnauthorized, "后台访问令牌格式不正确")
			c.Abort()
			return
		}

		if tokenString != adminBearerToken {
			response.ResponseError(c, http.StatusUnauthorized, "无效的后台访问令牌")
			c.Abort()
			return
		}

		c.Next()
	}
}
