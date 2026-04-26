package service

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/model"
)

const (
	taskStatusPending    = "pending"
	taskStatusProcessing = "processing"
)

type electricityTaskService struct {
	db               *gorm.DB
	buildingResolver BuildingResolver
	now              func() time.Time
}

func NewElectricityTaskService(db *gorm.DB, buildingResolver BuildingResolver, now func() time.Time) ElectricityTaskService {
	if now == nil {
		now = time.Now
	}

	return &electricityTaskService{
		db:               db,
		buildingResolver: buildingResolver,
		now:              now,
	}
}

func (s *electricityTaskService) Sync(userID uuid.UUID, req dto.SyncElectricityTaskRequest) error {
	for _, task := range req.Tasks {
		if _, err := s.buildingResolver.GetBuildingByCampusName(task.Campus, task.Building); err != nil {
			return ErrInvalidBuilding
		}

		if _, err := time.Parse("15:04", task.NotifyTime); err != nil {
			return ErrInvalidNotifyTime
		}
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var deviceToken model.DeviceToken
		err := tx.Where(model.DeviceToken{Token: req.DeviceToken}).
			Assign(model.DeviceToken{UserID: userID}).
			FirstOrCreate(&deviceToken).Error
		if err != nil {
			return err
		}

		var existingTasks []model.ElectricityTask
		if err := tx.Where("device_token_id = ?", deviceToken.ID).Find(&existingTasks).Error; err != nil {
			return err
		}

		incomingMap := make(map[string]dto.ElectricityTaskOption, len(req.Tasks))
		for _, task := range req.Tasks {
			incomingMap[electricityTaskKey(task.NotifyTime, task.Campus, task.Building, task.Room)] = task
		}

		existingMap := make(map[string]model.ElectricityTask, len(existingTasks))
		for _, task := range existingTasks {
			existingMap[electricityTaskKey(task.NotifyTime, task.Campus, task.Building, task.Room)] = task
		}

		for key, task := range existingMap {
			if _, ok := incomingMap[key]; ok {
				continue
			}
			if err := tx.Delete(&task).Error; err != nil {
				return err
			}
		}

		now := s.now()
		for key, task := range incomingMap {
			if _, ok := existingMap[key]; ok {
				continue
			}

			notifyTimeParsed, _ := time.Parse("15:04", task.NotifyTime)
			nextRunAt := time.Date(
				now.Year(), now.Month(), now.Day(),
				notifyTimeParsed.Hour(), notifyTimeParsed.Minute(), 0, 0, now.Location(),
			)
			if now.After(nextRunAt) {
				nextRunAt = nextRunAt.Add(24 * time.Hour)
			}

			newTask := model.ElectricityTask{
				DeviceTokenID: deviceToken.ID,
				NotifyTime:    task.NotifyTime,
				NextRunAt:     nextRunAt,
				Status:        taskStatusPending,
				Campus:        task.Campus,
				Building:      task.Building,
				Room:          task.Room,
			}
			if err := tx.Create(&newTask).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func electricityTaskKey(notifyTime, campus, building, room string) string {
	return notifyTime + "|" + campus + "|" + building + "|" + room
}
