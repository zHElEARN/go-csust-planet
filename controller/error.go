package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zHElEARN/go-csust-planet/utils"
)

func HandleNotFound(c *gin.Context) {
	utils.ResponseError(c, http.StatusNotFound, "route not found")
}
