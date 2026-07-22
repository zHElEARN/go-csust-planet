package router

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

const businessRequestTimeout = 2 * time.Second

func requestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
