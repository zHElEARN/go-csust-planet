package appversion

import (
	"context"

	"github.com/google/uuid"
)

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
func (s *Service) List(ctx context.Context) ([]Entity, error) {
	return s.repository.List(ctx)
}
func (s *Service) ListByPlatform(ctx context.Context, platform string) ([]Entity, error) {
	return s.repository.ListByPlatform(ctx, platform)
}
func (s *Service) Get(ctx context.Context, id uuid.UUID) (Entity, error) {
	return s.repository.Get(ctx, id)
}

func (s *Service) CheckUpdate(ctx context.Context, platform string, versionCode int) (CheckResult, error) {
	latest, err := s.repository.LatestByPlatform(ctx, platform)
	if err != nil || latest == nil {
		return CheckResult{LatestVersion: latest}, err
	}
	result := CheckResult{HasUpdate: latest.VersionCode > versionCode, LatestVersion: latest}
	if !result.HasUpdate {
		return result, nil
	}
	result.IsForceUpdate, err = s.repository.HasForceUpdateAfter(ctx, platform, versionCode)
	return result, err
}

func (s *Service) Create(ctx context.Context, input Upsert) (Entity, error) {
	return s.repository.Create(ctx, fromUpsert(input))
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input Upsert) (Entity, error) {
	return s.repository.Update(ctx, id, fromUpsert(input))
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}

func fromUpsert(input Upsert) Entity {
	return Entity{
		Platform: input.Platform, VersionCode: input.VersionCode, VersionName: input.VersionName,
		IsForceUpdate: input.IsForceUpdate, ReleaseNotes: input.ReleaseNotes, DownloadURL: input.DownloadURL,
	}
}
