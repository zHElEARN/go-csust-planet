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
		IsActive: true, IsBanner: true, CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("create announcement: %v", err)
	}

	updated, err := repository.Update(ctx, created.ID, Entity{
		Title: "after", Content: "after", IsActive: false, IsBanner: false,
	})
	if err != nil {
		t.Fatalf("update announcement: %v", err)
	}
	if updated.ID != created.ID || updated.Title != "after" || updated.IsActive || updated.IsBanner || !updated.CreatedAt.Equal(createdAt) {
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

func TestPostgresRepositoryHonorsCanceledContext(t *testing.T) {
	db, _ := testsupport.OpenTestDB(t, true)
	repository := NewPostgresRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := repository.List(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
