package campusmap

import "context"

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }
func (s *Service) List(ctx context.Context) ([]Entity, error) {
	return s.repository.List(ctx)
}
