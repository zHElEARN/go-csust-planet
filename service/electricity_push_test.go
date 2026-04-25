package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sideshow/apns2"

	"github.com/zHElEARN/go-csust-planet/model"
	"github.com/zHElEARN/go-csust-planet/utils/apns"
	"github.com/zHElEARN/go-csust-planet/utils/campuscard"
)

func TestElectricityPushServicePollAndProcessSuccess(t *testing.T) {
	db := openServiceTestDB(t)
	now := time.Date(2026, time.April, 26, 8, 0, 0, 0, time.UTC)

	user := createServiceTestUser(t, db, "20240003")
	deviceToken := createServiceTestDeviceToken(t, db, user.ID, "apns-token-success")
	task := createServiceTestElectricityTask(t, db, model.ElectricityTask{
		DeviceTokenID: deviceToken.ID,
		NotifyTime:    "07:00",
		NextRunAt:     now,
		Status:        taskStatusPending,
		Campus:        "云塘",
		Building:      "至诚轩1栋",
		Room:          "101",
	})

	var pushed apns.PushNotification
	pushService := NewElectricityPushService(
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
		ElectricityFetcherFunc(func(building campuscard.Building, roomNum string) (float64, error) {
			return 12.5, nil
		}),
		NotificationSenderFunc(func(notification apns.PushNotification) error {
			pushed = notification
			return nil
		}),
		ElectricityPushConfig{
			ZombieTaskThreshold: time.Minute,
			BatchSizeLimit:      10,
			TaskTimeout:         time.Second,
		},
	)

	if err := pushService.PollAndProcess(now); err != nil {
		t.Fatalf("expected poll to succeed: %v", err)
	}

	var updated model.ElectricityTask
	if err := db.First(&updated, "id = ?", task.ID).Error; err != nil {
		t.Fatalf("expected task to remain after success: %v", err)
	}
	expectedNextRunAt := time.Date(2026, time.April, 27, 7, 0, 0, 0, time.UTC)
	if updated.Status != taskStatusPending || !updated.NextRunAt.Equal(expectedNextRunAt) {
		t.Fatalf("unexpected updated task: %+v", updated)
	}
	if pushed.DeviceToken != deviceToken.Token || pushed.Title == "" || pushed.Body == "" {
		t.Fatalf("unexpected notification payload: %+v", pushed)
	}
}

func TestElectricityPushServiceDeletesTaskWhenRoomIsMissing(t *testing.T) {
	db := openServiceTestDB(t)
	now := time.Date(2026, time.April, 26, 8, 0, 0, 0, time.UTC)

	user := createServiceTestUser(t, db, "20240004")
	deviceToken := createServiceTestDeviceToken(t, db, user.ID, "apns-token-room-missing")
	task := createServiceTestElectricityTask(t, db, model.ElectricityTask{
		DeviceTokenID: deviceToken.ID,
		NotifyTime:    "07:00",
		NextRunAt:     now,
		Status:        taskStatusPending,
		Campus:        "云塘",
		Building:      "至诚轩1栋",
		Room:          "102",
	})

	pushService := NewElectricityPushService(
		db,
		BuildingResolverFunc(func(campusName, buildingName string) (campuscard.Building, error) {
			return campuscard.Building{Name: buildingName}, nil
		}),
		ElectricityFetcherFunc(func(building campuscard.Building, roomNum string) (float64, error) {
			return 0, campuscard.ErrRoomNotFound
		}),
		NotificationSenderFunc(func(notification apns.PushNotification) error {
			return nil
		}),
		ElectricityPushConfig{TaskTimeout: time.Second},
	)

	if err := pushService.PollAndProcess(now); err != nil {
		t.Fatalf("expected poll to succeed: %v", err)
	}

	var count int64
	if err := db.Model(&model.ElectricityTask{}).Where("id = ?", task.ID).Count(&count).Error; err != nil {
		t.Fatalf("failed to count tasks: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected missing-room task to be deleted, got %d rows", count)
	}
}

func TestElectricityPushServiceDeletesDeviceTokenOnBadAPNsToken(t *testing.T) {
	db := openServiceTestDB(t)
	now := time.Date(2026, time.April, 26, 8, 0, 0, 0, time.UTC)

	user := createServiceTestUser(t, db, "20240005")
	deviceToken := createServiceTestDeviceToken(t, db, user.ID, "apns-token-bad-device")
	createServiceTestElectricityTask(t, db, model.ElectricityTask{
		DeviceTokenID: deviceToken.ID,
		NotifyTime:    "07:00",
		NextRunAt:     now,
		Status:        taskStatusPending,
		Campus:        "云塘",
		Building:      "至诚轩1栋",
		Room:          "103",
	})

	pushService := NewElectricityPushService(
		db,
		BuildingResolverFunc(func(campusName, buildingName string) (campuscard.Building, error) {
			return campuscard.Building{Name: buildingName}, nil
		}),
		ElectricityFetcherFunc(func(building campuscard.Building, roomNum string) (float64, error) {
			return 9.9, nil
		}),
		NotificationSenderFunc(func(notification apns.PushNotification) error {
			return errors.New(apns2.ReasonBadDeviceToken)
		}),
		ElectricityPushConfig{TaskTimeout: time.Second},
	)

	if err := pushService.PollAndProcess(now); err != nil {
		t.Fatalf("expected poll to succeed: %v", err)
	}

	var deviceTokenCount int64
	if err := db.Model(&model.DeviceToken{}).Where("id = ?", deviceToken.ID).Count(&deviceTokenCount).Error; err != nil {
		t.Fatalf("failed to count device tokens: %v", err)
	}
	if deviceTokenCount != 0 {
		t.Fatalf("expected invalid device token to be deleted, got %d rows", deviceTokenCount)
	}
}

func TestElectricityPushServiceResetsTaskToPendingOnTimeout(t *testing.T) {
	db := openServiceTestDB(t)
	now := time.Date(2026, time.April, 26, 8, 0, 0, 0, time.UTC)

	user := createServiceTestUser(t, db, "20240006")
	deviceToken := createServiceTestDeviceToken(t, db, user.ID, "apns-token-timeout")
	task := createServiceTestElectricityTask(t, db, model.ElectricityTask{
		ID:            uuid.New(),
		DeviceTokenID: deviceToken.ID,
		NotifyTime:    "07:00",
		NextRunAt:     now,
		Status:        taskStatusPending,
		Campus:        "云塘",
		Building:      "至诚轩1栋",
		Room:          "104",
	})

	pushService := NewElectricityPushService(
		db,
		BuildingResolverFunc(func(campusName, buildingName string) (campuscard.Building, error) {
			return campuscard.Building{Name: buildingName}, nil
		}),
		ElectricityFetcherFunc(func(building campuscard.Building, roomNum string) (float64, error) {
			time.Sleep(50 * time.Millisecond)
			return 8.8, nil
		}),
		NotificationSenderFunc(func(notification apns.PushNotification) error {
			return nil
		}),
		ElectricityPushConfig{TaskTimeout: 10 * time.Millisecond},
	)

	if err := pushService.PollAndProcess(now); err != nil {
		t.Fatalf("expected poll to succeed: %v", err)
	}

	var updated model.ElectricityTask
	if err := db.First(&updated, "id = ?", task.ID).Error; err != nil {
		t.Fatalf("expected timed out task to remain: %v", err)
	}
	if updated.Status != taskStatusPending {
		t.Fatalf("expected timed out task to reset to pending, got %+v", updated)
	}
}
