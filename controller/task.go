package controller

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/model"
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
func SyncElectricityTask(c *gin.Context) {
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
		response.ResponseError(c, http.StatusBadRequest, "无效请求参数: "+err.Error())
		return
	}

	// 开启事务
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		var deviceToken model.DeviceToken
		// 查找或创建 DeviceToken
		err := tx.Where(model.DeviceToken{Token: req.DeviceToken}).
			Assign(model.DeviceToken{UserID: userID}).
			FirstOrCreate(&deviceToken).Error
		if err != nil {
			return err
		}

		// 加载现有任务
		var existingTasks []model.ElectricityTask
		if err := tx.Where("device_token_id = ?", deviceToken.ID).Find(&existingTasks).Error; err != nil {
			return err
		}

		// 用来快速比较的 map
		incomingMap := make(map[string]dto.ElectricityTaskOption)
		for _, t := range req.Tasks {
			key := t.NotifyTime + "|" + t.Campus + "|" + t.Building + "|" + t.Room
			incomingMap[key] = t
		}

		existingMap := make(map[string]model.ElectricityTask)
		for _, t := range existingTasks {
			key := t.NotifyTime + "|" + t.Campus + "|" + t.Building + "|" + t.Room
			existingMap[key] = t
		}

		// 删除不在 incoming 里面的任务
		for key, t := range existingMap {
			if _, ok := incomingMap[key]; !ok {
				if err := tx.Delete(&t).Error; err != nil {
					return err
				}
			}
		}

		// 增加在 incoming 里面有，但是在 existing 里面没有的任务
		now := time.Now()
		for key, t := range incomingMap {
			if _, ok := existingMap[key]; !ok {
				// 验证时间格式 "15:04"
				notifyTimeParsed, err := time.Parse("15:04", t.NotifyTime)
				if err != nil {
					return fmt.Errorf("notifyTime %s 格式错误，请使用 HH:mm 格式", t.NotifyTime)
				}

				nextRunAt := time.Date(now.Year(), now.Month(), now.Day(), notifyTimeParsed.Hour(), notifyTimeParsed.Minute(), 0, 0, now.Location())
				if now.After(nextRunAt) {
					nextRunAt = nextRunAt.Add(24 * time.Hour)
				}

				newTask := model.ElectricityTask{
					DeviceTokenID: deviceToken.ID,
					NotifyTime:    t.NotifyTime,
					NextRunAt:     nextRunAt,
					Status:        "pending",
					Campus:        t.Campus,
					Building:      t.Building,
					Room:          t.Room,
				}

				if err := tx.Create(&newTask).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("同步电费任务失败: %v\n", err)
		response.ResponseError(c, http.StatusInternalServerError, "同步任务失败: "+err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
