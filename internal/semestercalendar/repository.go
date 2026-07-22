package semestercalendar

type Repository interface {
	List() ([]Entity, error)
	ListSummaries() ([]Entity, error)
	Get(string) (Entity, error)
	Create(Entity) (Entity, error)
	Update(Entity) (Entity, error)
	Delete(string) error
}
