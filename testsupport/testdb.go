package testsupport

import (
	"context"
	"fmt"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	internalpostgres "github.com/zHElEARN/go-csust-planet/internal/postgres"
)

func OpenTestDB(t *testing.T, useTransaction bool) (*gorm.DB, *gorm.DB) {
	t.Helper()

	cfg, ok := LoadTestDBConfig()
	if !ok {
		t.Skip("skipping PostgreSQL integration test: set TEST_DB_HOST/TEST_DB_PORT/TEST_DB_USER/TEST_DB_PASSWORD/TEST_DB_NAME")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode, cfg.TimeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to obtain sql.DB from gorm: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if _, err := internalpostgres.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	if !useTransaction {
		return db, db
	}

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin transaction: %v", tx.Error)
	}
	t.Cleanup(func() {
		_ = tx.Rollback().Error
	})

	return tx, db
}
