package controller

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const healthCheckTimeout = 2 * time.Second

var errDatabaseUnavailable = errors.New("database unavailable")

func (h *Handler) HealthCheck(c *gin.Context) {
	if err := checkDatabaseHealth(c.Request.Context(), h.db); err != nil {
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

func checkDatabaseHealth(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return errDatabaseUnavailable
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	pingCtx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	return sqlDB.PingContext(pingCtx)
}
