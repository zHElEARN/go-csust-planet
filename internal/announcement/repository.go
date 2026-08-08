package announcement

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	List(context.Context) ([]Entity, error)
	ListActive(context.Context, string) ([]Entity, error)
	Get(context.Context, uuid.UUID) (Entity, error)
	Create(context.Context, Entity) (Entity, error)
	Update(context.Context, uuid.UUID, Entity) (Entity, error)
	Delete(context.Context, uuid.UUID) error
}
