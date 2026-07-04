package service

import (
	"github.com/google/uuid"

	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/model"
)

type AdminAppVersionService interface {
	List() ([]model.AppVersion, error)
	ListByPlatform(platform string) ([]model.AppVersion, error)
	Get(id uuid.UUID) (model.AppVersion, error)
	CheckUpdate(platform string, currentVersionCode int) (AppVersionCheckResult, error)
	Create(req dto.AdminAppVersionUpsertRequest) (model.AppVersion, error)
	Update(id uuid.UUID, req dto.AdminAppVersionUpsertRequest) (model.AppVersion, error)
	Delete(id uuid.UUID) error
}

type AppVersionCheckResult struct {
	HasUpdate     bool
	IsForceUpdate bool
	LatestVersion *model.AppVersion
}
