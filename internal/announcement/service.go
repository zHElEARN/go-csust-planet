package announcement

import (
	"time"

	"github.com/google/uuid"
)

type Service struct{ repository Repository }

type Upsert struct {
	Title    string
	Content  string
	IsActive bool
	IsBanner bool
}

func NewService(repository Repository) *Service     { return &Service{repository: repository} }
func (s *Service) List() ([]Entity, error)          { return s.repository.List() }
func (s *Service) ListActive() ([]Entity, error)    { return s.repository.ListActive() }
func (s *Service) Get(id uuid.UUID) (Entity, error) { return s.repository.Get(id) }

func (s *Service) Create(input Upsert) (Entity, error) {
	return s.repository.Create(Entity{ID: uuid.New(), Title: input.Title, Content: input.Content, IsActive: input.IsActive, IsBanner: input.IsBanner, CreatedAt: time.Now().UTC()})
}

func (s *Service) Update(id uuid.UUID, input Upsert) (Entity, error) {
	entity, err := s.repository.Get(id)
	if err != nil {
		return Entity{}, err
	}
	entity.Title, entity.Content = input.Title, input.Content
	entity.IsActive, entity.IsBanner = input.IsActive, input.IsBanner
	return s.repository.Update(entity)
}

func (s *Service) Delete(id uuid.UUID) error { return s.repository.Delete(id) }
