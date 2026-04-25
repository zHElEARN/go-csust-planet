package service

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/model"
)

type serviceTestDBConfig struct {
	host     string
	port     string
	user     string
	password string
	name     string
	sslMode  string
	timeZone string
}

var loadServiceTestEnvOnce sync.Once

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

	cfg, ok := loadServiceTestDBConfig()
	if !ok {
		t.Skip("skipping PostgreSQL service test: set TEST_DB_HOST/TEST_DB_PORT/TEST_DB_USER/TEST_DB_PASSWORD/TEST_DB_NAME")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.host, cfg.user, cfg.password, cfg.name, cfg.port, cfg.sslMode, cfg.timeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect service test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to obtain sql.DB from gorm: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
		t.Fatalf("failed to enable pgcrypto extension: %v", err)
	}
	if err := config.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate service test database: %v", err)
	}

	if !useTransaction {
		return db
	}

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin transaction: %v", tx.Error)
	}
	t.Cleanup(func() {
		_ = tx.Rollback().Error
	})

	return tx
}

func loadServiceTestDBConfig() (serviceTestDBConfig, bool) {
	loadServiceTestEnvOnce.Do(func() {
		_ = godotenv.Load(".env")
	})

	host := os.Getenv("TEST_DB_HOST")
	port := os.Getenv("TEST_DB_PORT")
	user := os.Getenv("TEST_DB_USER")
	password := os.Getenv("TEST_DB_PASSWORD")
	name := os.Getenv("TEST_DB_NAME")
	if host == "" || port == "" || user == "" || password == "" || name == "" {
		return serviceTestDBConfig{}, false
	}

	sslMode := os.Getenv("TEST_DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	timeZone := os.Getenv("TEST_DB_TIMEZONE")
	if timeZone == "" {
		timeZone = "Asia/Shanghai"
	}

	return serviceTestDBConfig{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		name:     name,
		sslMode:  sslMode,
		timeZone: timeZone,
	}, true
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
