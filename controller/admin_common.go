package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/zHElEARN/go-csust-planet/utils/response"
)

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的资源ID")
		return uuid.Nil, false
	}

	return id, true
}
