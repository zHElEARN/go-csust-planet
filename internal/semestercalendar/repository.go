package semestercalendar

import "context"

type Repository interface {
	List(context.Context) ([]Entity, error)
	ListSummaries(context.Context) ([]Entity, error)
	Get(context.Context, string) (Entity, error)
	Create(context.Context, Entity) (Entity, error)
	Update(context.Context, string, Entity) (Entity, error)
	Delete(context.Context, string) error
}
