package appversion

import "github.com/google/uuid"

type Service struct{ repository Repository }

type Upsert struct {
	Platform      string
	VersionCode   int
	VersionName   string
	IsForceUpdate bool
	ReleaseNotes  string
	DownloadURL   string
}

type CheckResult struct {
	HasUpdate     bool
	IsForceUpdate bool
	LatestVersion *Entity
}

func NewService(repository Repository) *Service { return &Service{repository: repository} }
func (s *Service) List() ([]Entity, error)      { return s.repository.List() }
func (s *Service) ListByPlatform(platform string) ([]Entity, error) {
	return s.repository.ListByPlatform(platform)
}
func (s *Service) Get(id uuid.UUID) (Entity, error) { return s.repository.Get(id) }

func (s *Service) CheckUpdate(platform string, versionCode int) (CheckResult, error) {
	latest, err := s.repository.LatestByPlatform(platform)
	if err != nil || latest == nil {
		return CheckResult{LatestVersion: latest}, err
	}
	result := CheckResult{HasUpdate: latest.VersionCode > versionCode, LatestVersion: latest}
	if !result.HasUpdate {
		return result, nil
	}
	result.IsForceUpdate, err = s.repository.HasForceUpdateAfter(platform, versionCode)
	return result, err
}

func (s *Service) Create(input Upsert) (Entity, error) {
	return s.repository.Create(Entity{Platform: input.Platform, VersionCode: input.VersionCode, VersionName: input.VersionName, IsForceUpdate: input.IsForceUpdate, ReleaseNotes: input.ReleaseNotes, DownloadURL: input.DownloadURL})
}

func (s *Service) Update(id uuid.UUID, input Upsert) (Entity, error) {
	entity, err := s.repository.Get(id)
	if err != nil {
		return Entity{}, err
	}
	entity.Platform, entity.VersionCode, entity.VersionName = input.Platform, input.VersionCode, input.VersionName
	entity.IsForceUpdate, entity.ReleaseNotes, entity.DownloadURL = input.IsForceUpdate, input.ReleaseNotes, input.DownloadURL
	return s.repository.Update(entity)
}

func (s *Service) Delete(id uuid.UUID) error { return s.repository.Delete(id) }
