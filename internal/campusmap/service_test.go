package campusmap

import "testing"

type fakeRepository struct {
	entities []Entity
	err      error
}

func (r *fakeRepository) List() ([]Entity, error) { return r.entities, r.err }

func TestServiceListsRepositoryEntities(t *testing.T) {
	expected := []Entity{{Type: "Feature"}}
	entities, err := NewService(&fakeRepository{entities: expected}).List()
	if err != nil || len(entities) != 1 || entities[0].Type != "Feature" {
		t.Fatalf("unexpected list result: %+v, %v", entities, err)
	}
}
