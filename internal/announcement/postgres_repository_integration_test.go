package announcement

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/zHElEARN/go-csust-planet/testsupport"
)

func TestPostgresRepositoryUpdateAndDeleteAreExplicit(t *testing.T) {
	db, _ := testsupport.OpenTestDB(t, true)
	repository := NewPostgresRepository(db)
	ctx := context.Background()

	createdAt := time.Now().UTC().Truncate(time.Microsecond)
	created, err := repository.Create(ctx, Entity{
		ID: uuid.New(), Title: "before", Content: "before",
		Platform: PlatformIOS, IsActive: true, IsBanner: true, CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("create announcement: %v", err)
	}

	updated, err := repository.Update(ctx, created.ID, Entity{
		Title: "after", Content: "after", Platform: PlatformAndroid, IsActive: false, IsBanner: false,
	})
	if err != nil {
		t.Fatalf("update announcement: %v", err)
	}
	if updated.ID != created.ID || updated.Title != "after" || updated.Platform != PlatformAndroid || updated.IsActive || updated.IsBanner || !updated.CreatedAt.Equal(createdAt) {
		t.Fatalf("unexpected returned announcement: %+v", updated)
	}

	missingID := uuid.New()
	if _, err := repository.Update(ctx, missingID, Entity{Title: "missing"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing update to return ErrNotFound, got %v", err)
	}
	var count int64
	if err := db.Model(&Entity{}).Where("id = ?", missingID).Count(&count).Error; err != nil {
		t.Fatalf("count missing announcement: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected missing update not to insert, got %d rows", count)
	}

	if err := repository.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete announcement: %v", err)
	}
	if err := repository.Delete(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected repeated delete to return ErrNotFound, got %v", err)
	}
}

func TestPostgresRepositoryListsActiveAnnouncementsByPlatform(t *testing.T) {
	db, _ := testsupport.OpenTestDB(t, true)
	repository := NewPostgresRepository(db)
	ctx := context.Background()

	fixtures := []Entity{
		{ID: uuid.New(), Title: "ios", Content: "ios", Platform: PlatformIOS, IsActive: true, CreatedAt: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), Title: "all", Content: "all", Platform: PlatformAll, IsActive: true, CreatedAt: time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), Title: "android", Content: "android", Platform: PlatformAndroid, IsActive: true, CreatedAt: time.Date(2026, time.January, 3, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), Title: "inactive", Content: "inactive", Platform: PlatformAll, IsActive: false, CreatedAt: time.Date(2026, time.January, 4, 0, 0, 0, 0, time.UTC)},
	}
	for _, fixture := range fixtures {
		if _, err := repository.Create(ctx, fixture); err != nil {
			t.Fatalf("create fixture %q: %v", fixture.Title, err)
		}
	}

	ios, err := repository.ListActive(ctx, PlatformIOS)
	if err != nil {
		t.Fatalf("list ios announcements: %v", err)
	}
	if len(ios) != 2 || ios[0].Platform != PlatformAll || ios[1].Platform != PlatformIOS {
		t.Fatalf("unexpected ios announcements: %+v", ios)
	}

	android, err := repository.ListActive(ctx, PlatformAndroid)
	if err != nil {
		t.Fatalf("list android announcements: %v", err)
	}
	if len(android) != 2 || android[0].Platform != PlatformAndroid || android[1].Platform != PlatformAll {
		t.Fatalf("unexpected android announcements: %+v", android)
	}
}

func TestPostgresRepositoryHonorsCanceledContext(t *testing.T) {
	db, _ := testsupport.OpenTestDB(t, true)
	repository := NewPostgresRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := repository.ListActive(ctx, PlatformIOS); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
