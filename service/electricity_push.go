package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sideshow/apns2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/zHElEARN/go-csust-planet/model"
	"github.com/zHElEARN/go-csust-planet/utils/apns"
	"github.com/zHElEARN/go-csust-planet/utils/campuscard"
)

const (
	DefaultWorkerTickInterval  = 1 * time.Minute
	defaultZombieTaskThreshold = 1 * time.Minute
	defaultBatchSizeLimit      = 100
	defaultTaskTimeout         = 30 * time.Second
)

type ElectricityPushConfig struct {
	ZombieTaskThreshold time.Duration
	BatchSizeLimit      int
	TaskTimeout         time.Duration
}

type electricityPushService struct {
	db               *gorm.DB
	buildingResolver BuildingResolver
	electricity      ElectricityFetcher
	notifier         NotificationSender
	config           ElectricityPushConfig
}

type taskWithToken struct {
	model.ElectricityTask
	Token string `gorm:"column:device_token"`
}

func DefaultElectricityPushConfig() ElectricityPushConfig {
	return ElectricityPushConfig{
		ZombieTaskThreshold: defaultZombieTaskThreshold,
		BatchSizeLimit:      defaultBatchSizeLimit,
		TaskTimeout:         defaultTaskTimeout,
	}
}

func NewElectricityPushService(
	db *gorm.DB,
	buildingResolver BuildingResolver,
	electricity ElectricityFetcher,
	notifier NotificationSender,
	cfg ElectricityPushConfig,
) ElectricityPushService {
	if cfg.ZombieTaskThreshold <= 0 {
		cfg.ZombieTaskThreshold = defaultZombieTaskThreshold
	}
	if cfg.BatchSizeLimit <= 0 {
		cfg.BatchSizeLimit = defaultBatchSizeLimit
	}
	if cfg.TaskTimeout <= 0 {
		cfg.TaskTimeout = defaultTaskTimeout
	}

	return &electricityPushService{
		db:               db,
		buildingResolver: buildingResolver,
		electricity:      electricity,
		notifier:         notifier,
		config:           cfg,
	}
}

func (s *electricityPushService) PollAndProcess(now time.Time) error {
	res := s.db.Model(&model.ElectricityTask{}).
		Where("status = ? AND updated_at <= ?", taskStatusProcessing, now.Add(-s.config.ZombieTaskThreshold)).
		Update("status", taskStatusPending)
	if res.Error != nil {
		return res.Error
	}

	tasks, err := s.claimPendingTasks(now)
	if err != nil {
		return err
	}

	deadTokens := make(map[string]bool)
	for _, task := range tasks {
		tokenIDStr := task.DeviceTokenID.String()
		if deadTokens[tokenIDStr] {
			continue
		}

		tokenInvalidated := s.processSingleTask(task, now)
		if tokenInvalidated {
			deadTokens[tokenIDStr] = true
		}
	}

	return nil
}

func (s *electricityPushService) claimPendingTasks(now time.Time) ([]taskWithToken, error) {
	var tasks []taskWithToken
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("electricity_tasks").
			Select("electricity_tasks.*, device_tokens.device_token").
			Joins("JOIN device_tokens on electricity_tasks.device_token_id = device_tokens.id").
			Where("electricity_tasks.status = ? AND electricity_tasks.next_run_at <= ?", taskStatusPending, now).
			Limit(s.config.BatchSizeLimit).
			Clauses(clause.Locking{
				Strength: "UPDATE",
				Table:    clause.Table{Name: "electricity_tasks"},
				Options:  "SKIP LOCKED",
			}).
			Scan(&tasks).Error; err != nil {
			return err
		}

		if len(tasks) == 0 {
			return nil
		}

		taskIDs := make([]string, 0, len(tasks))
		for _, task := range tasks {
			taskIDs = append(taskIDs, task.ID.String())
		}

		return tx.Model(&model.ElectricityTask{}).
			Where("id IN ?", taskIDs).
			Updates(map[string]any{
				"status":     taskStatusProcessing,
				"updated_at": now,
			}).Error
	})
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *electricityPushService) processSingleTask(task taskWithToken, batchStartTime time.Time) bool {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.TaskTimeout)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		electricityVal, err := s.fetchRealElectricity(task.Campus, task.Building, task.Room)
		if err != nil {
			errCh <- fmt.Errorf("获取电量失败: %w", err)
			return
		}

		notification := apns.PushNotification{
			DeviceToken: task.Token,
			Title:       "宿舍电量通知",
			Body:        fmt.Sprintf("%s%s宿舍当前电量: %s", task.Building, task.Room, electricityVal),
			Sound:       "default",
		}
		errCh <- s.notifier.SendPushNotification(notification)
	}()

	var taskErr error
	select {
	case <-ctx.Done():
		taskErr = fmt.Errorf("任务执行超时(%v)", s.config.TaskTimeout)
	case err := <-errCh:
		taskErr = err
	}

	if taskErr != nil {
		if errors.Is(taskErr, campuscard.ErrRoomNotFound) {
			_ = s.db.Delete(&model.ElectricityTask{}, "id = ?", task.ID).Error
			return false
		}

		reason := taskErr.Error()
		if reason == apns2.ReasonUnregistered || reason == apns2.ReasonBadDeviceToken {
			if err := s.db.Where("id = ?", task.DeviceTokenID).Delete(&model.DeviceToken{}).Error; err != nil {
				_ = s.resetTaskPending(task.ID)
				return false
			}
			return true
		}

		_ = s.resetTaskPending(task.ID)
		return false
	}

	notifyTimeParsed, err := time.Parse("15:04", task.NotifyTime)
	if err != nil {
		_ = s.resetTaskPending(task.ID)
		return false
	}

	nextRunAt := time.Date(
		batchStartTime.Year(), batchStartTime.Month(), batchStartTime.Day(),
		notifyTimeParsed.Hour(), notifyTimeParsed.Minute(), 0, 0, batchStartTime.Location(),
	)
	if !nextRunAt.After(batchStartTime) {
		nextRunAt = nextRunAt.Add(24 * time.Hour)
	}

	_ = s.db.Model(&model.ElectricityTask{}).
		Where("id = ?", task.ID).
		Updates(map[string]any{
			"next_run_at": nextRunAt,
			"status":      taskStatusPending,
			"updated_at":  time.Now(),
		}).Error

	return false
}

func (s *electricityPushService) fetchRealElectricity(campusName, buildingName, roomNum string) (string, error) {
	targetBuilding, err := s.buildingResolver.GetBuildingByCampusName(campusName, buildingName)
	if err != nil {
		return "", err
	}

	balance, err := s.electricity.GetElectricity(targetBuilding, roomNum)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", balance), nil
}

func (s *electricityPushService) resetTaskPending(taskID any) error {
	return s.db.Model(&model.ElectricityTask{}).
		Where("id = ?", taskID).
		Updates(map[string]any{
			"status":     taskStatusPending,
			"updated_at": time.Now(),
		}).Error
}
