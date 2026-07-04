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
