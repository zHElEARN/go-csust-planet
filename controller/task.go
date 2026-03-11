package controller

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/model"
	"github.com/zHElEARN/go-csust-planet/utils/response"
)

type addTaskRequest struct {
	DeviceToken string `json:"device_token" binding:"required"`
	NotifyTime  string `json:"notify_time" binding:"required"`
	Campus      string `json:"campus" binding:"required"`
	Building    string `json:"building" binding:"required"`
	Room        string `json:"room" binding:"required"`
}

// AddElectricityTask godoc
// @Summary      添加电费推送任务
// @Description  添加一个新的电费定时推送任务
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        request  body      addTaskRequest  true  "任务请求内容"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /task/electricity [post]
// @Security     BearerAuth
func AddElectricityTask(c *gin.Context) {
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

	var req addTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效请求参数: "+err.Error())
		return
	}

	// 验证时间格式 "15:04"
	_, err = time.Parse("15:04", req.NotifyTime)
	if err != nil {
		response.ResponseError(c, http.StatusBadRequest, "notify_time 格式错误，请使用 HH:mm 格式，例如 15:04")
		return
	}

	// 开启事务
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		var deviceToken model.DeviceToken
		// 查找或创建 DeviceToken
		err := tx.Where(model.DeviceToken{Token: req.DeviceToken, UserID: userID}).
			FirstOrCreate(&deviceToken).Error
		if err != nil {
			return err
		}

		// 计算下次执行时间
		now := time.Now()
		notifyTimeParsed, _ := time.Parse("15:04", req.NotifyTime)

		nextRunAt := time.Date(now.Year(), now.Month(), now.Day(), notifyTimeParsed.Hour(), notifyTimeParsed.Minute(), 0, 0, now.Location())
		if now.After(nextRunAt) {
			// 如果今天的时间已经过了，那就设置为明天
			nextRunAt = nextRunAt.Add(24 * time.Hour)
		}

		// 创建任务
		task := model.ElectricityTask{
			DeviceTokenID: deviceToken.ID,
			NotifyTime:    req.NotifyTime,
			NextRunAt:     nextRunAt,
			Status:        "pending",
			Campus:        req.Campus,
			Building:      req.Building,
			Room:          req.Room,
		}

		if err := tx.Create(&task).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("添加电费任务失败: %v\n", err)
		response.ResponseError(c, http.StatusInternalServerError, "添加任务失败: "+err.Error())
		return
	}

	response.ResponseSuccess(c, "电费任务添加成功")
}
