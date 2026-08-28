package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func corsMiddleware(origins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		allowed[origin] = struct{}{}
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if !strings.HasPrefix(path, "/v1") && !strings.HasPrefix(path, "/v2") {
			c.Next()
			return
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
