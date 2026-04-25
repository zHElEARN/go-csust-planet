package controller

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/service"
	"github.com/zHElEARN/go-csust-planet/utils/response"
)

// SyncElectricityTask godoc
// @Summary      同步电费推送任务
// @Description  同步电费定时推送任务
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        request  body      dto.SyncElectricityTaskRequest  true  "任务请求内容"
// @Success      204      "成功，无返回内容"
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /task/electricity [post]
// @Security     BearerAuth
func (h *Handler) SyncElectricityTask(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		response.ResponseError(c, http.StatusUnauthorized, "未授权的访问")
		return
	}

	userID, err := uuid.Parse(userIdStr.(string))
	if err != nil {
		response.ResponseError(c, http.StatusUnauthorized, "无效的用户ID")
		return
	}

	var req dto.SyncElectricityTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效请求参数")
		return
	}

	err = h.electricityTaskService.Sync(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidBuilding):
			response.ResponseError(c, http.StatusBadRequest, "无效的校区或楼栋")
		case errors.Is(err, service.ErrInvalidNotifyTime):
			response.ResponseError(c, http.StatusBadRequest, "notifyTime 格式错误，请使用 HH:mm 格式")
		default:
			log.Printf("[ERROR] 同步电费任务失败: %v", err)
			response.ResponseError(c, http.StatusInternalServerError, "同步任务失败")
		}
		return
	}

	c.Status(http.StatusNoContent)
}
