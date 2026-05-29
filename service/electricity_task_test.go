package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/model"
	"github.com/zHElEARN/go-csust-planet/utils/csustkit"
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
	var validatedRooms []string
	taskService := NewElectricityTaskService(
		db,
		ElectricityRoomValidatorFunc(func(ctx context.Context, campusName, buildingName, roomName string) error {
			validatedRooms = append(validatedRooms, roomName)
			return nil
		}),
		func() time.Time { return now },
	)

	err := taskService.Sync(context.Background(), user.ID, dto.SyncElectricityTaskRequest{
		DeviceToken: deviceToken.Token,
		Tasks: []dto.ElectricityTaskOption{
			{NotifyTime: "08:30", Campus: "云塘", Building: "至诚轩1栋", Room: "101"},
			{NotifyTime: "08:00", Campus: "云塘", Building: "至诚轩1栋", Room: "103"},
		},
	})
	if err != nil {
		t.Fatalf("expected sync to succeed: %v", err)
	}
	if len(validatedRooms) != 2 {
		t.Fatalf("expected 2 room validations, got %d", len(validatedRooms))
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
		ElectricityRoomValidatorFunc(func(ctx context.Context, campusName, buildingName, roomName string) error {
			return csustkit.ErrRoomNotFound
		}),
		nil,
	)

	err := taskService.Sync(context.Background(), uuid.New(), dto.SyncElectricityTaskRequest{
		DeviceToken: "device-token",
		Tasks: []dto.ElectricityTaskOption{
			{NotifyTime: "08:00", Campus: "云塘", Building: "未知楼栋", Room: "101"},
		},
	})
	if !errors.Is(err, ErrInvalidRoom) {
		t.Fatalf("expected ErrInvalidRoom, got %v", err)
	}
	var taskCount int64
	if err := db.Model(&model.ElectricityTask{}).Count(&taskCount).Error; err != nil {
		t.Fatalf("failed to count electricity tasks: %v", err)
	}
	if taskCount != 0 {
		t.Fatalf("expected invalid room to skip transaction writes, got %d tasks", taskCount)
	}

	upstreamErr := errors.New("campus card unavailable")
	taskService = NewElectricityTaskService(
		db,
		ElectricityRoomValidatorFunc(func(ctx context.Context, campusName, buildingName, roomName string) error {
			return upstreamErr
		}),
		nil,
	)
	err = taskService.Sync(context.Background(), uuid.New(), dto.SyncElectricityTaskRequest{
		DeviceToken: "device-token",
		Tasks: []dto.ElectricityTaskOption{
			{NotifyTime: "08:00", Campus: "云塘", Building: "至诚轩1栋", Room: "101"},
		},
	})
	if !errors.Is(err, upstreamErr) {
		t.Fatalf("expected upstream error, got %v", err)
	}

	var validatorCalled bool
	taskService = NewElectricityTaskService(
		db,
		ElectricityRoomValidatorFunc(func(ctx context.Context, campusName, buildingName, roomName string) error {
			validatorCalled = true
			return nil
		}),
		nil,
	)

	err = taskService.Sync(context.Background(), uuid.New(), dto.SyncElectricityTaskRequest{
		DeviceToken: "device-token",
		Tasks: []dto.ElectricityTaskOption{
			{NotifyTime: "8am", Campus: "云塘", Building: "至诚轩1栋", Room: "101"},
		},
	})
	if !errors.Is(err, ErrInvalidNotifyTime) {
		t.Fatalf("expected ErrInvalidNotifyTime, got %v", err)
	}
	if validatorCalled {
		t.Fatalf("expected invalid notify time to skip room validation")
	}
}
