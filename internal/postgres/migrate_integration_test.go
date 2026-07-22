package postgres_test

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	internalpostgres "github.com/zHElEARN/go-csust-planet/internal/postgres"
	"github.com/zHElEARN/go-csust-planet/testsupport"
)

type tableFingerprint struct {
	TableName   string
	RowCount    int64
	Fingerprint string
}

//go:embed migrations/00001_initial_schema.sql
var initialMigrationSQL string

func TestMigratePreservesBusinessDataAndIsIdempotent(t *testing.T) {
	db := openTestDB(t)

	var tableCount int64
	if err := db.Raw(`
		SELECT count(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		  AND table_name IN ('announcements', 'app_versions', 'campus_map_features', 'semester_calendars')
	`).Scan(&tableCount).Error; err != nil {
		t.Fatalf("failed to count business tables: %v", err)
	}
	if tableCount == 0 {
		version, err := internalpostgres.Migrate(context.Background(), db)
		if err != nil {
			t.Fatalf("fresh database migration failed: %v", err)
		}
		if version != 1 {
			t.Fatalf("fresh database migration returned version %d, want 1", version)
		}
	} else if tableCount != 4 {
		t.Fatalf("test database has %d of 4 business tables", tableCount)
	}

	before := readFingerprints(t, db)
	for run := 1; run <= 2; run++ {
		version, err := internalpostgres.Migrate(context.Background(), db)
		if err != nil {
			t.Fatalf("migration run %d failed: %v", run, err)
		}
		if version != 1 {
			t.Fatalf("migration run %d returned version %d, want 1", run, version)
		}
	}
	after := readFingerprints(t, db)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("business data changed during migration: before=%v after=%v", before, after)
	}
}

func TestInitialMigrationAcceptsCurrentSchemaReadOnly(t *testing.T) {
	db := openTestDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to obtain sql database: %v", err)
	}

	const statementEnd = "-- +goose StatementEnd"
	start := strings.Index(initialMigrationSQL, "DO $migration$")
	end := strings.Index(initialMigrationSQL, statementEnd)
	if start < 0 || end <= start {
		t.Fatal("failed to locate V1 baseline validation block")
	}

	tx, err := sqlDB.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatalf("failed to begin read-only schema validation: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })
	if _, err := tx.ExecContext(context.Background(), initialMigrationSQL[start:end]); err != nil {
		t.Fatalf("current schema is incompatible with V1: %v", err)
	}
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	cfg, ok := testsupport.LoadTestDBConfig()
	if !ok {
		t.Skip("skipping PostgreSQL migration test: set TEST_DB_HOST/TEST_DB_PORT/TEST_DB_USER/TEST_DB_PASSWORD/TEST_DB_NAME")
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
		t.Fatalf("failed to obtain sql database: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}

func readFingerprints(t *testing.T, db *gorm.DB) []tableFingerprint {
	t.Helper()

	const query = `
		SELECT 'announcements' AS table_name, count(*) AS row_count,
		       md5(COALESCE(string_agg(md5(row_to_json(t)::text), '' ORDER BY id), '')) AS fingerprint
		FROM public.announcements AS t
		UNION ALL
		SELECT 'app_versions', count(*),
		       md5(COALESCE(string_agg(md5(row_to_json(t)::text), '' ORDER BY id), ''))
		FROM public.app_versions AS t
		UNION ALL
		SELECT 'campus_map_features', count(*),
		       md5(COALESCE(string_agg(md5(row_to_json(t)::text), '' ORDER BY id), ''))
		FROM public.campus_map_features AS t
		UNION ALL
		SELECT 'semester_calendars', count(*),
		       md5(COALESCE(string_agg(md5(row_to_json(t)::text), '' ORDER BY id), ''))
		FROM public.semester_calendars AS t
		ORDER BY table_name`

	var fingerprints []tableFingerprint
	if err := db.Raw(query).Scan(&fingerprints).Error; err != nil {
		t.Fatalf("failed to fingerprint business data: %v", err)
	}
	return fingerprints
}
