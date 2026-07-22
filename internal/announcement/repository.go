package announcement

import "github.com/google/uuid"

type Repository interface {
	List() ([]Entity, error)
	ListActive() ([]Entity, error)
	Get(uuid.UUID) (Entity, error)
	Create(Entity) (Entity, error)
	Update(Entity) (Entity, error)
	Delete(uuid.UUID) error
}
