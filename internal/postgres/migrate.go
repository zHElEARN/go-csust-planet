package postgres

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Migrate(ctx context.Context, db *gorm.DB) (int64, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return 0, fmt.Errorf("get sql database: %w", err)
	}

	migrations, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return 0, fmt.Errorf("open embedded migrations: %w", err)
	}

	locker, err := lock.NewPostgresSessionLocker(
		lock.WithLockTimeout(1, 60),
		lock.WithUnlockTimeout(1, 10),
	)
	if err != nil {
		return 0, fmt.Errorf("create migration lock: %w", err)
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		sqlDB,
		migrations,
		goose.WithSessionLocker(locker),
	)
	if err != nil {
		return 0, fmt.Errorf("create migration provider: %w", err)
	}

	if _, err := provider.Up(ctx); err != nil {
		return 0, fmt.Errorf("apply migrations: %w", err)
	}

	version, err := provider.GetDBVersion(ctx)
	if err != nil {
		return 0, fmt.Errorf("read migration version: %w", err)
	}
	return version, nil
}
