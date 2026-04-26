package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zHElEARN/go-csust-planet/utils/jwt"
	"github.com/zHElEARN/go-csust-planet/utils/response"
)

// AuthMiddleware 身份验证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := parseBearerToken(c.GetHeader("Authorization"))
		if errors.Is(err, errAuthorizationHeaderMissing) {
			response.ResponseError(c, http.StatusUnauthorized, "未提供身份验证令牌")
			c.Abort()
			return
		}
		if errors.Is(err, errAuthorizationHeaderInvalid) {
			response.ResponseError(c, http.StatusUnauthorized, "身份验证令牌格式不正确")
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(tokenString)
		if err != nil {
			response.ResponseError(c, http.StatusUnauthorized, "无效或过期的令牌")
			c.Abort()
			return
		}

		// 将 userID 存储在上下文中
		c.Set("userID", claims.Subject)
		c.Next()
	}
}
