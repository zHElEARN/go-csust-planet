package service

import (
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/model"
	"github.com/zHElEARN/go-csust-planet/testsupport"
)

func openServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	return openServiceTestDBWithCleanup(t, true)
}

func openPersistentServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	return openServiceTestDBWithCleanup(t, false)
}

func openServiceTestDBWithCleanup(t *testing.T, useTransaction bool) *gorm.DB {
	t.Helper()

	testDB, _ := testsupport.OpenTestDB(t, useTransaction)
	return testDB
}
func createServiceTestUser(t *testing.T, db *gorm.DB, studentID string) model.User {
	t.Helper()

	user := model.User{
		ID:        uuid.New(),
		StudentID: studentID,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return user
}

func createServiceTestDeviceToken(t *testing.T, db *gorm.DB, userID uuid.UUID, token string) model.DeviceToken {
	t.Helper()

	deviceToken := model.DeviceToken{
		ID:     uuid.New(),
		Token:  token,
		UserID: userID,
	}
	if err := db.Create(&deviceToken).Error; err != nil {
		t.Fatalf("failed to create device token: %v", err)
	}

	return deviceToken
}

func createServiceTestElectricityTask(t *testing.T, db *gorm.DB, task model.ElectricityTask) model.ElectricityTask {
	t.Helper()

	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}
	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("failed to create electricity task: %v", err)
	}

	return task
}

func createServiceTestAppVersion(t *testing.T, db *gorm.DB, version model.AppVersion) model.AppVersion {
	t.Helper()

	if version.ID == uuid.Nil {
		version.ID = uuid.New()
	}
	if err := db.Create(&version).Error; err != nil {
		t.Fatalf("failed to create app version: %v", err)
	}

	return version
}
