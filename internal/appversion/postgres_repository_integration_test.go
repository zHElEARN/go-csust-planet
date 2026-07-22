package appversion

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/zHElEARN/go-csust-planet/testsupport"
)

func TestPostgresRepositoryEnforcesUniquenessAndExplicitUpdates(t *testing.T) {
	db, _ := testsupport.OpenTestDB(t, true)
	repository := NewPostgresRepository(db)
	ctx := context.Background()
	baseCode := int(time.Now().UnixNano() % 1_000_000_000)

	created, err := repository.Create(ctx, Entity{
		Platform: "ios", VersionCode: baseCode, VersionName: "before",
		IsForceUpdate: true, ReleaseNotes: "before", DownloadURL: "https://example.com/before",
	})
	if err != nil {
		t.Fatalf("create app version: %v", err)
	}
	if _, err := repository.Create(ctx, Entity{
		Platform: "ios", VersionCode: baseCode, VersionName: "duplicate",
		ReleaseNotes: "duplicate", DownloadURL: "https://example.com/duplicate",
	}); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected duplicate create to return ErrConflict, got %v", err)
	}

	updated, err := repository.Update(ctx, created.ID, Entity{
		Platform: "android", VersionCode: baseCode + 1, VersionName: "after",
		IsForceUpdate: false, ReleaseNotes: "after", DownloadURL: "https://example.com/after",
	})
	if err != nil {
		t.Fatalf("update app version: %v", err)
	}
	if updated.ID != created.ID || updated.Platform != "android" || updated.VersionCode != baseCode+1 || updated.IsForceUpdate || updated.CreatedAt.IsZero() {
		t.Fatalf("unexpected returned app version: %+v", updated)
	}

	conflicting, err := repository.Create(ctx, Entity{
		Platform: "ios", VersionCode: baseCode + 2, VersionName: "conflicting",
		ReleaseNotes: "conflicting", DownloadURL: "https://example.com/conflicting",
	})
	if err != nil {
		t.Fatalf("create conflicting app version fixture: %v", err)
	}
	if _, err := repository.Update(ctx, conflicting.ID, Entity{
		Platform: updated.Platform, VersionCode: updated.VersionCode, VersionName: "duplicate update",
		ReleaseNotes: "duplicate update", DownloadURL: "https://example.com/duplicate-update",
	}); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected duplicate update to return ErrConflict, got %v", err)
	}

	missingID := uuid.New()
	if _, err := repository.Update(ctx, missingID, Entity{Platform: "ios", VersionCode: baseCode + 3}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing update to return ErrNotFound, got %v", err)
	}
	var count int64
	if err := db.Model(&Entity{}).Where("id = ?", missingID).Count(&count).Error; err != nil {
		t.Fatalf("count missing app version: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected missing update not to insert, got %d rows", count)
	}

	if err := repository.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete app version: %v", err)
	}
	if err := repository.Delete(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected repeated delete to return ErrNotFound, got %v", err)
	}
}

func TestPostgresRepositoryHonorsCanceledContext(t *testing.T) {
	db, _ := testsupport.OpenTestDB(t, true)
	repository := NewPostgresRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := repository.List(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
