package announcement

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	entity  Entity
	getErr  error
	created Entity
	updated Entity
	deleted uuid.UUID
}

func (r *fakeRepository) List() ([]Entity, error)       { return nil, nil }
func (r *fakeRepository) ListActive() ([]Entity, error) { return nil, nil }
func (r *fakeRepository) Get(uuid.UUID) (Entity, error) { return r.entity, r.getErr }
func (r *fakeRepository) Create(entity Entity) (Entity, error) {
	r.created = entity
	return entity, nil
}
func (r *fakeRepository) Update(entity Entity) (Entity, error) {
	r.updated = entity
	return entity, nil
}
func (r *fakeRepository) Delete(id uuid.UUID) error { r.deleted = id; return nil }

func TestServiceCreatesAndUpdatesAnnouncement(t *testing.T) {
	repository := &fakeRepository{entity: Entity{ID: uuid.New(), Title: "old"}}
	service := NewService(repository)
	created, err := service.Create(Upsert{Title: "new", Content: "content", IsActive: true, IsBanner: true})
	if err != nil || created.ID == uuid.Nil || repository.created.Title != "new" {
		t.Fatalf("unexpected create result: %+v, %v", created, err)
	}
	updated, err := service.Update(repository.entity.ID, Upsert{Title: "updated", Content: "body", IsActive: false, IsBanner: true})
	if err != nil || updated.Title != "updated" || repository.updated.IsActive {
		t.Fatalf("unexpected update result: %+v, %v", updated, err)
	}
}

func TestServiceReturnsRepositoryNotFound(t *testing.T) {
	service := NewService(&fakeRepository{getErr: ErrNotFound})
	_, err := service.Update(uuid.New(), Upsert{})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
