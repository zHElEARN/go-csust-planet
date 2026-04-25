package controller

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zHElEARN/go-csust-planet/config"
)

const healthCheckTimeout = 2 * time.Second

var errDatabaseUnavailable = errors.New("database unavailable")

func HealthCheck(c *gin.Context) {
	if err := checkDatabaseHealth(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "unavailable",
			"database": "down",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func checkDatabaseHealth(ctx context.Context) error {
	if config.DB == nil {
		return errDatabaseUnavailable
	}

	sqlDB, err := config.DB.DB()
	if err != nil {
		return err
	}

	pingCtx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	return sqlDB.PingContext(pingCtx)
}
