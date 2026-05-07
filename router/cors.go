package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/zHElEARN/go-csust-planet/config"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/v1") {
			c.Next()
			return
		}

		if config.AppConfig == nil {
			c.Next()
			return
		}

		allowed := make(map[string]struct{}, len(config.AppConfig.CORSAllowedOrigins))
		for _, origin := range config.AppConfig.CORSAllowedOrigins {
			allowed[origin] = struct{}{}
		}

		origin := c.GetHeader("Origin")
		if origin != "" {
			if _, ok := allowed[origin]; ok {
				headers := c.Writer.Header()
				headers.Set("Access-Control-Allow-Origin", origin)
				headers.Add("Vary", "Origin")
				headers.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				headers.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				headers.Set("Access-Control-Max-Age", "86400")
			}
		}

		if c.Request.Method == http.MethodOptions {
			if _, ok := allowed[origin]; ok {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		}

		c.Next()
	}
}
