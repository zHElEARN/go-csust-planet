package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zHElEARN/go-csust-planet/utils/response"
)

func HandleNotFound(c *gin.Context) {
	response.ResponseError(c, http.StatusNotFound, "route not found")
}
