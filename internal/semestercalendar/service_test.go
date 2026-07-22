package semestercalendar

import (
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	entity  Entity
	getErr  error
	created Entity
	updated Entity
}

func (r *fakeRepository) List() ([]Entity, error)          { return nil, nil }
func (r *fakeRepository) ListSummaries() ([]Entity, error) { return nil, nil }
func (r *fakeRepository) Get(string) (Entity, error)       { return r.entity, r.getErr }
func (r *fakeRepository) Create(entity Entity) (Entity, error) {
	r.created = entity
	return entity, nil
}
func (r *fakeRepository) Update(entity Entity) (Entity, error) {
	r.updated = entity
	return entity, nil
}
func (r *fakeRepository) Delete(string) error { return nil }

func TestServiceRenamesCalendarAndNormalizesCollections(t *testing.T) {
	repository := &fakeRepository{entity: Entity{SemesterCode: "2024-2025-1"}}
	service := NewService(repository)
	updated, err := service.Update("2024-2025-1", Upsert{SemesterCode: "2024-2025-1A", Title: "title", Subtitle: "subtitle", CalendarStart: time.Now(), CalendarEnd: time.Now(), SemesterStart: time.Now(), SemesterEnd: time.Now()})
	if err != nil || updated.SemesterCode != "2024-2025-1A" || repository.updated.Notes == nil || repository.updated.CustomWeekRanges == nil {
		t.Fatalf("unexpected update result: %+v, %v", updated, err)
	}
}

func TestServiceReturnsNotFound(t *testing.T) {
	service := NewService(&fakeRepository{getErr: ErrNotFound})
	_, err := service.Update("missing", Upsert{})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
