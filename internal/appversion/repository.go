package appversion

import "github.com/google/uuid"

type Repository interface {
	List() ([]Entity, error)
	ListByPlatform(string) ([]Entity, error)
	Get(uuid.UUID) (Entity, error)
	LatestByPlatform(string) (*Entity, error)
	HasForceUpdateAfter(string, int) (bool, error)
	Create(Entity) (Entity, error)
	Update(Entity) (Entity, error)
	Delete(uuid.UUID) error
}
