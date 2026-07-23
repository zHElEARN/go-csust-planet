package router

import (
	"net/http"
	"time"

	timeout "github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"

	"github.com/zHElEARN/go-csust-planet/utils/response"
)

func requestTimeout(duration time.Duration) gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(duration),
		timeout.WithResponse(func(c *gin.Context) {
			response.ResponseError(c, http.StatusServiceUnavailable, "请求处理超时")
		}),
	)
}
