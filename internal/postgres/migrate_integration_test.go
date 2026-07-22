package postgres_test

import (
	"context"
	_ "embed"
	"fmt"
	"reflect"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/internal/appversion"
	internalpostgres "github.com/zHElEARN/go-csust-planet/internal/postgres"
	"github.com/zHElEARN/go-csust-planet/testsupport"
)

type tableFingerprint struct {
	TableName   string
	RowCount    int64
	Fingerprint string
}

//go:embed migrations/00002_unique_app_version.sql
var uniqueAppVersionMigrationSQL string

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
		if version != 2 {
			t.Fatalf("fresh database migration returned version %d, want 2", version)
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
		if version != 2 {
			t.Fatalf("migration run %d returned version %d, want 2", run, version)
		}
	}
	after := readFingerprints(t, db)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("business data changed during migration: before=%v after=%v", before, after)
	}
}

func TestAppVersionPlatformVersionIndexIsUnique(t *testing.T) {
	db := openTestDB(t)

	var unique bool
	if err := db.Raw(`
		SELECT i.indisunique
		FROM pg_catalog.pg_index AS i
		JOIN pg_catalog.pg_class AS c ON c.oid = i.indexrelid
		JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public' AND c.relname = 'idx_platform_version'
	`).Scan(&unique).Error; err != nil {
		t.Fatalf("failed to inspect app version index: %v", err)
	}
	if !unique {
		t.Fatal("expected idx_platform_version to be unique")
	}

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin duplicate check transaction: %v", tx.Error)
	}
	t.Cleanup(func() { _ = tx.Rollback().Error })

	versionCode := int(time.Now().UnixNano() % 1_000_000_000)
	first := appversion.Entity{Platform: "migration-test", VersionCode: versionCode, VersionName: "first", ReleaseNotes: "first", DownloadURL: "https://example.com/first"}
	second := appversion.Entity{Platform: first.Platform, VersionCode: versionCode, VersionName: "second", ReleaseNotes: "second", DownloadURL: "https://example.com/second"}
	if err := tx.Create(&first).Error; err != nil {
		t.Fatalf("failed to create first app version: %v", err)
	}
	if err := tx.Create(&second).Error; err == nil {
		t.Fatal("expected duplicate app version insert to fail")
	}
}

func TestUniqueAppVersionMigrationRejectsDuplicatesWithoutDeletingThem(t *testing.T) {
	db := openTestDB(t)
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin migration failure transaction: %v", tx.Error)
	}
	t.Cleanup(func() { _ = tx.Rollback().Error })

	if err := tx.Exec("DROP INDEX public.idx_platform_version").Error; err != nil {
		t.Fatalf("failed to drop unique index fixture: %v", err)
	}
	if err := tx.Exec("CREATE INDEX idx_platform_version ON public.app_versions (platform, version_code)").Error; err != nil {
		t.Fatalf("failed to create non-unique index fixture: %v", err)
	}

	versionCode := int(time.Now().UnixNano() % 1_000_000_000)
	for _, name := range []string{"first", "second"} {
		entity := appversion.Entity{
			Platform: "duplicate-migration-test", VersionCode: versionCode, VersionName: name,
			ReleaseNotes: name, DownloadURL: "https://example.com/" + name,
		}
		if err := tx.Create(&entity).Error; err != nil {
			t.Fatalf("failed to create duplicate fixture %q: %v", name, err)
		}
	}

	if err := tx.Exec("SAVEPOINT before_v2_migration").Error; err != nil {
		t.Fatalf("failed to create migration savepoint: %v", err)
	}
	if err := tx.Exec(uniqueAppVersionMigrationSQL).Error; err == nil {
		t.Fatal("expected V2 migration to fail for duplicate app versions")
	}
	if err := tx.Exec("ROLLBACK TO SAVEPOINT before_v2_migration").Error; err != nil {
		t.Fatalf("failed to recover from expected migration error: %v", err)
	}

	var count int64
	if err := tx.Model(&appversion.Entity{}).
		Where("platform = ? AND version_code = ?", "duplicate-migration-test", versionCode).
		Count(&count).Error; err != nil {
		t.Fatalf("failed to count duplicate fixtures: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected migration failure to preserve both duplicate rows, got %d", count)
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
