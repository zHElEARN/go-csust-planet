package semestercalendar

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	entity    Entity
	getErr    error
	updateErr error
	created   Entity
	updated   Entity
	updateCtx context.Context
	getCalls  int
}

func (r *fakeRepository) List(context.Context) ([]Entity, error)          { return nil, nil }
func (r *fakeRepository) ListSummaries(context.Context) ([]Entity, error) { return nil, nil }
func (r *fakeRepository) Get(context.Context, string) (Entity, error) {
	r.getCalls++
	return r.entity, r.getErr
}
func (r *fakeRepository) Create(_ context.Context, entity Entity) (Entity, error) {
	r.created = entity
	return entity, nil
}
func (r *fakeRepository) Update(ctx context.Context, _ string, entity Entity) (Entity, error) {
	r.updateCtx = ctx
	r.updated = entity
	return entity, r.updateErr
}
func (r *fakeRepository) Delete(context.Context, string) error { return nil }

func TestServiceRenamesCalendarAndNormalizesCollections(t *testing.T) {
	repository := &fakeRepository{entity: Entity{SemesterCode: "2024-2025-1"}}
	service := NewService(repository)
	ctx := context.WithValue(context.Background(), struct{}{}, "request")
	updated, err := service.Update(ctx, "2024-2025-1", Upsert{SemesterCode: "2024-2025-1A", Title: "title", Subtitle: "subtitle", CalendarStart: time.Now(), CalendarEnd: time.Now(), SemesterStart: time.Now(), SemesterEnd: time.Now()})
	if err != nil || updated.SemesterCode != "2024-2025-1A" || repository.updated.Notes == nil || repository.updated.CustomWeekRanges == nil {
		t.Fatalf("unexpected update result: %+v, %v", updated, err)
	}
	if repository.updateCtx != ctx || repository.getCalls != 0 {
		t.Fatal("expected context propagation without a pre-read")
	}
}

func TestServiceReturnsNotFound(t *testing.T) {
	service := NewService(&fakeRepository{updateErr: ErrNotFound})
	_, err := service.Update(context.Background(), "missing", Upsert{})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
