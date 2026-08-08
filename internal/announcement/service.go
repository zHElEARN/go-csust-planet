package announcement

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Service struct{ repository Repository }

type Upsert struct {
	Title    string
	Content  string
	Platform string
	IsActive bool
	IsBanner bool
}

func NewService(repository Repository) *Service { return &Service{repository: repository} }
func (s *Service) List(ctx context.Context) ([]Entity, error) {
	return s.repository.List(ctx)
}
func (s *Service) ListActive(ctx context.Context, platform string) ([]Entity, error) {
	return s.repository.ListActive(ctx, platform)
}
func (s *Service) Get(ctx context.Context, id uuid.UUID) (Entity, error) {
	return s.repository.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input Upsert) (Entity, error) {
	return s.repository.Create(ctx, Entity{ID: uuid.New(), Title: input.Title, Content: input.Content, Platform: input.Platform, IsActive: input.IsActive, IsBanner: input.IsBanner, CreatedAt: time.Now().UTC()})
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input Upsert) (Entity, error) {
	return s.repository.Update(ctx, id, Entity{
		Title: input.Title, Content: input.Content, Platform: input.Platform,
		IsActive: input.IsActive, IsBanner: input.IsBanner,
	})
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}
