package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/model"
	"github.com/zHElEARN/go-csust-planet/utils/campuscard"
)

func TestElectricityTaskServiceSyncDiffsTasksAndSchedulesNewOnes(t *testing.T) {
	db := openServiceTestDB(t)
	user := createServiceTestUser(t, db, "20240002")
	deviceToken := createServiceTestDeviceToken(t, db, user.ID, "device-token-1")

	keptTask := createServiceTestElectricityTask(t, db, model.ElectricityTask{
		DeviceTokenID: deviceToken.ID,
		NotifyTime:    "08:30",
		NextRunAt:     time.Date(2026, time.April, 25, 8, 30, 0, 0, time.UTC),
		Status:        taskStatusPending,
		Campus:        "云塘",
		Building:      "至诚轩1栋",
		Room:          "101",
	})
	createServiceTestElectricityTask(t, db, model.ElectricityTask{
		DeviceTokenID: deviceToken.ID,
		NotifyTime:    "07:30",
		NextRunAt:     time.Date(2026, time.April, 25, 7, 30, 0, 0, time.UTC),
		Status:        taskStatusPending,
		Campus:        "云塘",
		Building:      "至诚轩1栋",
		Room:          "102",
	})

	now := time.Date(2026, time.April, 26, 9, 0, 0, 0, time.UTC)
	taskService := NewElectricityTaskService(
		db,
		BuildingResolverFunc(func(campusName, buildingName string) (campuscard.Building, error) {
			return campuscard.Building{
				ID:   "building-1",
				Name: buildingName,
				Campus: campuscard.Campus{
					ID:          "campus-1",
					DisplayName: campusName,
				},
			}, nil
		}),
		func() time.Time { return now },
	)

	err := taskService.Sync(user.ID, dto.SyncElectricityTaskRequest{
		DeviceToken: deviceToken.Token,
		Tasks: []dto.ElectricityTaskOption{
			{NotifyTime: "08:30", Campus: "云塘", Building: "至诚轩1栋", Room: "101"},
			{NotifyTime: "08:00", Campus: "云塘", Building: "至诚轩1栋", Room: "103"},
		},
	})
	if err != nil {
		t.Fatalf("expected sync to succeed: %v", err)
	}

	var tasks []model.ElectricityTask
	if err := db.Where("device_token_id = ?", deviceToken.ID).Order("notify_time asc, room asc").Find(&tasks).Error; err != nil {
		t.Fatalf("failed to query synced tasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks after sync, got %d", len(tasks))
	}

	var keptFound bool
	var newFound bool
	expectedNextRunAt := time.Date(2026, time.April, 27, 8, 0, 0, 0, time.UTC)
	for _, task := range tasks {
		switch task.Room {
		case keptTask.Room:
			keptFound = true
			if task.ID != keptTask.ID {
				t.Fatalf("expected existing task to be preserved, got %+v", task)
			}
		case "103":
			newFound = true
			if !task.NextRunAt.Equal(expectedNextRunAt) {
				t.Fatalf("expected new task next_run_at %s, got %s", expectedNextRunAt, task.NextRunAt)
			}
			if task.Status != taskStatusPending {
				t.Fatalf("expected new task status pending, got %s", task.Status)
			}
		}
	}
	if !keptFound || !newFound {
		t.Fatalf("expected kept and new tasks to exist, got %+v", tasks)
	}
}

func TestElectricityTaskServiceSyncValidatesInput(t *testing.T) {
	db := openServiceTestDB(t)
	taskService := NewElectricityTaskService(
		db,
		BuildingResolverFunc(func(campusName, buildingName string) (campuscard.Building, error) {
			return campuscard.Building{}, errors.New("invalid building")
		}),
		nil,
	)

	err := taskService.Sync(uuid.New(), dto.SyncElectricityTaskRequest{
		DeviceToken: "device-token",
		Tasks: []dto.ElectricityTaskOption{
			{NotifyTime: "08:00", Campus: "云塘", Building: "未知楼栋", Room: "101"},
		},
	})
	if !errors.Is(err, ErrInvalidBuilding) {
		t.Fatalf("expected ErrInvalidBuilding, got %v", err)
	}

	taskService = NewElectricityTaskService(
		db,
		BuildingResolverFunc(func(campusName, buildingName string) (campuscard.Building, error) {
			return campuscard.Building{}, nil
		}),
		nil,
	)

	err = taskService.Sync(uuid.New(), dto.SyncElectricityTaskRequest{
		DeviceToken: "device-token",
		Tasks: []dto.ElectricityTaskOption{
			{NotifyTime: "8am", Campus: "云塘", Building: "至诚轩1栋", Room: "101"},
		},
	})
	if !errors.Is(err, ErrInvalidNotifyTime) {
		t.Fatalf("expected ErrInvalidNotifyTime, got %v", err)
	}
}
