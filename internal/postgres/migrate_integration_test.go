package postgres_test

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
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

//go:embed migrations/00003_announcement_platform.sql
var announcementPlatformMigrationSQL string

//go:embed migrations/00004_split_app_versions.sql
var splitAppVersionsMigrationSQL string

func TestMigratePreservesBusinessDataAndIsIdempotent(t *testing.T) {
	db := openTestDB(t)

	version, err := internalpostgres.Migrate(context.Background(), db)
	if err != nil {
		t.Fatalf("fresh database migration failed: %v", err)
	}
	if version != 4 {
		t.Fatalf("fresh database migration returned version %d, want 4", version)
	}

	var tableCount int64
	if err := db.Raw(`
		SELECT count(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		  AND table_name IN ('announcements', 'legacy_app_versions', 'app_versions', 'campus_map_features', 'semester_calendars')
	`).Scan(&tableCount).Error; err != nil {
		t.Fatalf("failed to count business tables: %v", err)
	}
	if tableCount != 5 {
		t.Fatalf("test database has %d of 5 business tables", tableCount)
	}

	before := readFingerprints(t, db)
	for run := 1; run <= 2; run++ {
		version, err := internalpostgres.Migrate(context.Background(), db)
		if err != nil {
			t.Fatalf("migration run %d failed: %v", run, err)
		}
		if version != 4 {
			t.Fatalf("migration run %d returned version %d, want 4", run, version)
		}
	}
	after := readFingerprints(t, db)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("business data changed during migration: before=%v after=%v", before, after)
	}
}

func TestAppVersionSplitMigrationPreservesLegacyAndCreatesEmptyCurrent(t *testing.T) {
	db := openTestDB(t)
	if _, err := internalpostgres.Migrate(context.Background(), db); err != nil {
		t.Fatalf("prepare migrated database: %v", err)
	}

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin app version split transaction: %v", tx.Error)
	}
	t.Cleanup(func() { _ = tx.Rollback().Error })

	if err := tx.Exec("DROP TABLE public.app_versions").Error; err != nil {
		t.Fatalf("failed to remove current app version fixture table: %v", err)
	}
	if err := tx.Exec("ALTER TABLE public.legacy_app_versions RENAME TO app_versions").Error; err != nil {
		t.Fatalf("failed to restore v3 app version table name: %v", err)
	}
	if err := tx.Exec("ALTER TABLE public.app_versions RENAME CONSTRAINT legacy_app_versions_pkey TO app_versions_pkey").Error; err != nil {
		t.Fatalf("failed to restore v3 app version primary key name: %v", err)
	}
	if err := tx.Exec("ALTER INDEX public.idx_legacy_platform_version RENAME TO idx_platform_version").Error; err != nil {
		t.Fatalf("failed to restore v3 app version index name: %v", err)
	}

	legacyID := uuid.New()
	if err := tx.Exec(`
		INSERT INTO public.app_versions (
			id, platform, version_code, version_name, is_force_update,
			release_notes, download_url, created_at
		) VALUES (?, 'android', 123456789, 'migration-ready', true, 'export data', 'https://example.com/legacy.apk', CURRENT_TIMESTAMP)
	`, legacyID).Error; err != nil {
		t.Fatalf("failed to create v3 app version fixture: %v", err)
	}

	var beforeCount int64
	if err := tx.Table(appversion.TableName).Count(&beforeCount).Error; err != nil {
		t.Fatalf("failed to count v3 app versions: %v", err)
	}
	if err := tx.Exec(splitAppVersionsMigrationSQL).Error; err != nil {
		t.Fatalf("app version split migration failed: %v", err)
	}

	var legacyCount int64
	if err := tx.Table(appversion.LegacyTableName).Count(&legacyCount).Error; err != nil {
		t.Fatalf("failed to count migrated legacy app versions: %v", err)
	}
	if legacyCount != beforeCount {
		t.Fatalf("legacy row count changed during split: got %d, want %d", legacyCount, beforeCount)
	}

	var migrated appversion.Entity
	if err := tx.Table(appversion.LegacyTableName).First(&migrated, "id = ?", legacyID).Error; err != nil {
		t.Fatalf("failed to read migrated legacy app version: %v", err)
	}
	if migrated.VersionName != "migration-ready" || !migrated.IsForceUpdate || migrated.DownloadURL != "https://example.com/legacy.apk" {
		t.Fatalf("legacy app version changed during split: %+v", migrated)
	}

	var currentCount int64
	if err := tx.Table(appversion.TableName).Count(&currentCount).Error; err != nil {
		t.Fatalf("failed to count current app versions: %v", err)
	}
	if currentCount != 0 {
		t.Fatalf("expected new app_versions table to be empty, got %d rows", currentCount)
	}

	for _, indexName := range []string{"idx_legacy_platform_version", "idx_platform_version"} {
		var unique bool
		if err := tx.Raw(`
			SELECT i.indisunique
			FROM pg_catalog.pg_index AS i
			JOIN pg_catalog.pg_class AS c ON c.oid = i.indexrelid
			JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
			WHERE n.nspname = 'public' AND c.relname = ?
		`, indexName).Scan(&unique).Error; err != nil {
			t.Fatalf("failed to inspect %s: %v", indexName, err)
		}
		if !unique {
			t.Fatalf("expected %s to be unique", indexName)
		}
	}

	current := appversion.Entity{
		Platform: "android", VersionCode: 123456789, VersionName: "new-package",
		ReleaseNotes: "new package", DownloadURL: "https://example.com/current.apk",
	}
	if err := tx.Table(appversion.TableName).Create(&current).Error; err != nil {
		t.Fatalf("same platform and version code should be allowed across tables: %v", err)
	}
}

func TestAnnouncementPlatformMigrationBackfillsAndConstrains(t *testing.T) {
	db := openTestDB(t)
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin announcement migration transaction: %v", tx.Error)
	}
	t.Cleanup(func() { _ = tx.Rollback().Error })

	if err := tx.Exec("ALTER TABLE public.announcements DROP CONSTRAINT announcements_platform_check").Error; err != nil {
		t.Fatalf("failed to drop platform check fixture: %v", err)
	}
	if err := tx.Exec("DROP INDEX public.idx_active_platform_created").Error; err != nil {
		t.Fatalf("failed to drop platform index fixture: %v", err)
	}
	if err := tx.Exec("ALTER TABLE public.announcements DROP COLUMN platform").Error; err != nil {
		t.Fatalf("failed to drop platform column fixture: %v", err)
	}
	if err := tx.Exec("CREATE INDEX idx_active_created ON public.announcements (is_active, created_at)").Error; err != nil {
		t.Fatalf("failed to create legacy announcement index fixture: %v", err)
	}

	legacyID := uuid.New()
	if err := tx.Exec(`
		INSERT INTO public.announcements (id, title, content, is_active, is_banner, created_at)
		VALUES (?, ?, ?, true, false, CURRENT_TIMESTAMP)
	`, legacyID, "legacy ios announcement", "legacy content").Error; err != nil {
		t.Fatalf("failed to create legacy announcement fixture: %v", err)
	}
	if err := tx.Exec(announcementPlatformMigrationSQL).Error; err != nil {
		t.Fatalf("announcement platform migration failed: %v", err)
	}

	var platform string
	if err := tx.Raw("SELECT platform FROM public.announcements WHERE id = ?", legacyID).Scan(&platform).Error; err != nil {
		t.Fatalf("failed to read migrated announcement: %v", err)
	}
	if platform != "ios" {
		t.Fatalf("migrated announcement platform = %q, want ios", platform)
	}

	var column struct {
		IsNullable    string
		ColumnDefault sql.NullString
	}
	if err := tx.Raw(`
		SELECT is_nullable, column_default
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND table_name = 'announcements'
		  AND column_name = 'platform'
	`).Scan(&column).Error; err != nil {
		t.Fatalf("failed to inspect announcement platform column: %v", err)
	}
	if column.IsNullable != "NO" || column.ColumnDefault.Valid {
		t.Fatalf("unexpected platform column definition: %+v", column)
	}

	if err := tx.Exec("SAVEPOINT before_invalid_announcement_platform").Error; err != nil {
		t.Fatalf("failed to create invalid platform savepoint: %v", err)
	}
	invalidErr := tx.Exec(`
		INSERT INTO public.announcements (id, title, content, platform, is_active, is_banner, created_at)
		VALUES (?, ?, ?, 'windows', true, false, CURRENT_TIMESTAMP)
	`, uuid.New(), "invalid platform", "invalid content").Error
	if invalidErr == nil {
		t.Fatal("expected invalid announcement platform insert to fail")
	}
	if err := tx.Exec("ROLLBACK TO SAVEPOINT before_invalid_announcement_platform").Error; err != nil {
		t.Fatalf("failed to recover from invalid platform error: %v", err)
	}

	var indexDefinition string
	if err := tx.Raw(`
		SELECT indexdef
		FROM pg_catalog.pg_indexes
		WHERE schemaname = 'public' AND indexname = 'idx_active_platform_created'
	`).Scan(&indexDefinition).Error; err != nil {
		t.Fatalf("failed to inspect announcement platform index: %v", err)
	}
	wantIndex := "CREATE INDEX idx_active_platform_created ON public.announcements USING btree (is_active, platform, created_at DESC)"
	if indexDefinition != wantIndex {
		t.Fatalf("unexpected announcement platform index: %q", indexDefinition)
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
		       md5(COALESCE(string_agg(md5(jsonb_build_object(
		           'id', t.id,
		           'title', t.title,
		           'content', t.content,
		           'is_active', t.is_active,
		           'is_banner', t.is_banner,
		           'created_at', t.created_at
		       )::text), '' ORDER BY id), '')) AS fingerprint
		FROM public.announcements AS t
		UNION ALL
		SELECT 'app_versions', count(*),
		       md5(COALESCE(string_agg(md5(row_to_json(t)::text), '' ORDER BY id), ''))
		FROM public.app_versions AS t
		UNION ALL
		SELECT 'legacy_app_versions', count(*),
		       md5(COALESCE(string_agg(md5(row_to_json(t)::text), '' ORDER BY id), ''))
		FROM public.legacy_app_versions AS t
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
