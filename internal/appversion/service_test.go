package appversion

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	entity  Entity
	latest  *Entity
	getErr  error
	force   bool
	created Entity
	updated Entity
}

func (r *fakeRepository) List() ([]Entity, error)                       { return nil, nil }
func (r *fakeRepository) ListByPlatform(string) ([]Entity, error)       { return nil, nil }
func (r *fakeRepository) Get(uuid.UUID) (Entity, error)                 { return r.entity, r.getErr }
func (r *fakeRepository) LatestByPlatform(string) (*Entity, error)      { return r.latest, nil }
func (r *fakeRepository) HasForceUpdateAfter(string, int) (bool, error) { return r.force, nil }
func (r *fakeRepository) Create(entity Entity) (Entity, error) {
	r.created = entity
	return entity, nil
}
func (r *fakeRepository) Update(entity Entity) (Entity, error) {
	r.updated = entity
	return entity, nil
}
func (r *fakeRepository) Delete(uuid.UUID) error { return nil }

func TestServiceChecksForcedUpdate(t *testing.T) {
	latest := Entity{Platform: "ios", VersionCode: 200}
	service := NewService(&fakeRepository{latest: &latest, force: true})
	result, err := service.CheckUpdate("ios", 100)
	if err != nil || !result.HasUpdate || !result.IsForceUpdate || result.LatestVersion != &latest {
		t.Fatalf("unexpected result: %+v, %v", result, err)
	}
}

func TestServiceUpdatesExistingVersion(t *testing.T) {
	repository := &fakeRepository{entity: Entity{ID: uuid.New(), Platform: "ios", VersionCode: 1}}
	service := NewService(repository)
	updated, err := service.Update(repository.entity.ID, Upsert{Platform: "android", VersionCode: 2, VersionName: "2.0", IsForceUpdate: true, ReleaseNotes: "notes", DownloadURL: "https://example.com"})
	if err != nil || updated.Platform != "android" || repository.updated.VersionCode != 2 {
		t.Fatalf("unexpected update result: %+v, %v", updated, err)
	}
}

func TestServiceReturnsNotFound(t *testing.T) {
	service := NewService(&fakeRepository{getErr: ErrNotFound})
	_, err := service.Update(uuid.New(), Upsert{})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
