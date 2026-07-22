package semestercalendar

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/zHElEARN/go-csust-planet/testsupport"
)

func TestPostgresRepositoryUpdatesCollectionsAndDeletesExplicitly(t *testing.T) {
	db, _ := testsupport.OpenTestDB(t, true)
	repository := NewPostgresRepository(db)
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	code := fmt.Sprintf("repository-%d", suffix)
	created, err := repository.Create(ctx, Entity{
		SemesterCode: code, Title: "before", Subtitle: "before",
		CalendarStart:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		CalendarEnd:      time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		SemesterStart:    time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		SemesterEnd:      time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC),
		Notes:            []CalendarNote{{Row: 1, Content: "before"}},
		CustomWeekRanges: []CustomWeekRange{{StartRow: 1, EndRow: 2, Content: "before"}},
	})
	if err != nil {
		t.Fatalf("create semester calendar: %v", err)
	}

	renamedCode := code + "-renamed"
	updated, err := repository.Update(ctx, code, Entity{
		SemesterCode: renamedCode, Title: "after", Subtitle: "after",
		CalendarStart: created.CalendarStart, CalendarEnd: created.CalendarEnd,
		SemesterStart: created.SemesterStart, SemesterEnd: created.SemesterEnd,
		Notes: []CalendarNote{}, CustomWeekRanges: []CustomWeekRange{},
	})
	if err != nil {
		t.Fatalf("update semester calendar: %v", err)
	}
	if updated.ID != created.ID || updated.SemesterCode != renamedCode || updated.Title != "after" || updated.CreatedAt.IsZero() {
		t.Fatalf("unexpected returned semester calendar: %+v", updated)
	}
	if updated.Notes == nil || len(updated.Notes) != 0 || updated.CustomWeekRanges == nil || len(updated.CustomWeekRanges) != 0 {
		t.Fatalf("expected returned collections to be empty arrays: %+v %+v", updated.Notes, updated.CustomWeekRanges)
	}

	missingCode := fmt.Sprintf("missing-%d", suffix)
	if _, err := repository.Update(ctx, missingCode, Entity{
		SemesterCode: missingCode, Title: "missing", Subtitle: "missing",
		CalendarStart: created.CalendarStart, CalendarEnd: created.CalendarEnd,
		SemesterStart: created.SemesterStart, SemesterEnd: created.SemesterEnd,
		Notes: []CalendarNote{}, CustomWeekRanges: []CustomWeekRange{},
	}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing update to return ErrNotFound, got %v", err)
	}
	var count int64
	if err := db.Model(&Entity{}).Where("semester_code = ?", missingCode).Count(&count).Error; err != nil {
		t.Fatalf("count missing semester calendar: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected missing update not to insert, got %d rows", count)
	}

	if err := repository.Delete(ctx, renamedCode); err != nil {
		t.Fatalf("delete semester calendar: %v", err)
	}
	if err := repository.Delete(ctx, renamedCode); !errors.Is(err, ErrNotFound) {
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
