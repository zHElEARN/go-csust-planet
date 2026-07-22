package campusmap

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }
func (s *Service) List() ([]Entity, error)      { return s.repository.List() }
