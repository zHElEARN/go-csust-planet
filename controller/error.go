package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"code":    404,
		"message": "route not found",
	})
}
