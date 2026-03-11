package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zHElEARN/go-csust-planet/utils"
)

// AuthMiddleware 身份验证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ResponseError(c, http.StatusUnauthorized, "未提供身份验证令牌")
			c.Abort()
			return
		}

		// 按空格分割，通常是 "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			utils.ResponseError(c, http.StatusUnauthorized, "身份验证令牌格式不正确")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			utils.ResponseError(c, http.StatusUnauthorized, "无效或过期的令牌")
			c.Abort()
			return
		}

		// 将 userID 存储在上下文中，供后续处理程序使用
		c.Set("userID", claims.Subject)
		c.Next()
	}
}
