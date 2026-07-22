package health

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const timeout = 2 * time.Second

type Handler struct{ db *gorm.DB }

func NewHandler(db *gorm.DB) *Handler { return &Handler{db: db} }

func (h *Handler) Check(c *gin.Context) {
	if err := checkDatabase(c.Request.Context(), h.db); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "database": "down"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func checkDatabase(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return errors.New("database unavailable")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return sqlDB.PingContext(pingCtx)
}
