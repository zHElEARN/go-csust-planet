package appversion

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	entity    Entity
	latest    *Entity
	getErr    error
	updateErr error
	force     bool
	created   Entity
	updated   Entity
	latestCtx context.Context
	forceCtx  context.Context
	updateCtx context.Context
	getCalls  int
}

func (r *fakeRepository) List(context.Context) ([]Entity, error)                   { return nil, nil }
func (r *fakeRepository) ListByPlatform(context.Context, string) ([]Entity, error) { return nil, nil }
func (r *fakeRepository) Get(context.Context, uuid.UUID) (Entity, error) {
	r.getCalls++
	return r.entity, r.getErr
}
func (r *fakeRepository) LatestByPlatform(ctx context.Context, _ string) (*Entity, error) {
	r.latestCtx = ctx
	return r.latest, nil
}
func (r *fakeRepository) HasForceUpdateAfter(ctx context.Context, _ string, _ int) (bool, error) {
	r.forceCtx = ctx
	return r.force, nil
}
func (r *fakeRepository) Create(_ context.Context, entity Entity) (Entity, error) {
	r.created = entity
	return entity, nil
}
func (r *fakeRepository) Update(ctx context.Context, _ uuid.UUID, entity Entity) (Entity, error) {
	r.updateCtx = ctx
	r.updated = entity
	return entity, r.updateErr
}
func (r *fakeRepository) Delete(context.Context, uuid.UUID) error { return nil }

func TestServiceChecksForcedUpdate(t *testing.T) {
	latest := Entity{Platform: "ios", VersionCode: 200}
	repository := &fakeRepository{latest: &latest, force: true}
	service := NewService(repository)
	ctx := context.WithValue(context.Background(), struct{}{}, "request")
	result, err := service.CheckUpdate(ctx, "ios", 100)
	if err != nil || !result.HasUpdate || !result.IsForceUpdate || result.LatestVersion != &latest {
		t.Fatalf("unexpected result: %+v, %v", result, err)
	}
	if repository.latestCtx != ctx || repository.forceCtx != ctx {
		t.Fatal("expected context to reach both update-check queries")
	}
}

func TestServiceUpdatesExistingVersion(t *testing.T) {
	repository := &fakeRepository{entity: Entity{ID: uuid.New(), Platform: "ios", VersionCode: 1}}
	service := NewService(repository)
	ctx := context.WithValue(context.Background(), struct{}{}, "request")
	updated, err := service.Update(ctx, repository.entity.ID, Upsert{Platform: "android", VersionCode: 2, VersionName: "2.0", IsForceUpdate: true, ReleaseNotes: "notes", DownloadURL: "https://example.com"})
	if err != nil || updated.Platform != "android" || repository.updated.VersionCode != 2 {
		t.Fatalf("unexpected update result: %+v, %v", updated, err)
	}
	if repository.updateCtx != ctx || repository.getCalls != 0 {
		t.Fatal("expected context propagation without a pre-read")
	}
}

func TestServiceReturnsNotFound(t *testing.T) {
	service := NewService(&fakeRepository{updateErr: ErrNotFound})
	_, err := service.Update(context.Background(), uuid.New(), Upsert{})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
