package announcement

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	entity    Entity
	getErr    error
	updateErr error
	created   Entity
	updated   Entity
	deleted   uuid.UUID
	createCtx context.Context
	updateCtx context.Context
	getCalls  int
}

func (r *fakeRepository) List(context.Context) ([]Entity, error)       { return nil, nil }
func (r *fakeRepository) ListActive(context.Context) ([]Entity, error) { return nil, nil }
func (r *fakeRepository) Get(context.Context, uuid.UUID) (Entity, error) {
	r.getCalls++
	return r.entity, r.getErr
}
func (r *fakeRepository) Create(ctx context.Context, entity Entity) (Entity, error) {
	r.createCtx = ctx
	r.created = entity
	return entity, nil
}
func (r *fakeRepository) Update(ctx context.Context, _ uuid.UUID, entity Entity) (Entity, error) {
	r.updateCtx = ctx
	r.updated = entity
	return entity, r.updateErr
}
func (r *fakeRepository) Delete(_ context.Context, id uuid.UUID) error { r.deleted = id; return nil }

func TestServiceCreatesAndUpdatesAnnouncement(t *testing.T) {
	repository := &fakeRepository{entity: Entity{ID: uuid.New(), Title: "old"}}
	service := NewService(repository)
	ctx := context.WithValue(context.Background(), struct{}{}, "request")
	created, err := service.Create(ctx, Upsert{Title: "new", Content: "content", IsActive: true, IsBanner: true})
	if err != nil || created.ID == uuid.Nil || repository.created.Title != "new" {
		t.Fatalf("unexpected create result: %+v, %v", created, err)
	}
	updated, err := service.Update(ctx, repository.entity.ID, Upsert{Title: "updated", Content: "body", IsActive: false, IsBanner: true})
	if err != nil || updated.Title != "updated" || repository.updated.IsActive {
		t.Fatalf("unexpected update result: %+v, %v", updated, err)
	}
	if repository.createCtx != ctx || repository.updateCtx != ctx || repository.getCalls != 0 {
		t.Fatalf("expected context propagation without a pre-read")
	}
}

func TestServiceReturnsRepositoryNotFound(t *testing.T) {
	service := NewService(&fakeRepository{updateErr: ErrNotFound})
	_, err := service.Update(context.Background(), uuid.New(), Upsert{})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
