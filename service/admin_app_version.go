package service

import (
	"errors"
	"hash/crc32"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/model"
)

type adminAppVersionService struct {
	db *gorm.DB
}

func NewAdminAppVersionService(db *gorm.DB) AdminAppVersionService {
	return &adminAppVersionService{db: db}
}

func (s *adminAppVersionService) List() ([]model.AppVersion, error) {
	var versions []model.AppVersion
	if err := s.db.Order("platform asc, version_code desc").Find(&versions).Error; err != nil {
		return nil, err
	}

	return versions, nil
}

func (s *adminAppVersionService) ListByPlatform(platform string) ([]model.AppVersion, error) {
	var versions []model.AppVersion
	if err := s.db.Where("platform = ?", platform).Order("version_code desc").Find(&versions).Error; err != nil {
		return nil, err
	}

	return versions, nil
}

func (s *adminAppVersionService) Get(id uuid.UUID) (model.AppVersion, error) {
	var version model.AppVersion
	if err := s.db.First(&version, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.AppVersion{}, ErrNotFound
		}
		return model.AppVersion{}, err
	}

	return version, nil
}

func (s *adminAppVersionService) CheckUpdate(platform string, currentVersionCode int) (AppVersionCheckResult, error) {
	var latestVersion model.AppVersion
	if err := s.db.Where("platform = ?", platform).Order("version_code desc").First(&latestVersion).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AppVersionCheckResult{}, nil
		}
		return AppVersionCheckResult{}, err
	}

	result := AppVersionCheckResult{
		HasUpdate:     latestVersion.VersionCode > currentVersionCode,
		IsForceUpdate: false,
		LatestVersion: &latestVersion,
	}
	if !result.HasUpdate {
		return result, nil
	}

	var forceUpdate model.AppVersion
	err := s.db.Select("id").
		Where("platform = ? AND version_code > ? AND is_force_update = ?", platform, currentVersionCode, true).
		First(&forceUpdate).Error
	if err == nil {
		result.IsForceUpdate = true
		return result, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return result, nil
	}

	return AppVersionCheckResult{}, err
}

func (s *adminAppVersionService) Create(req dto.AdminAppVersionUpsertRequest) (model.AppVersion, error) {
	var version model.AppVersion
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := lockAppVersion(tx, req.Platform, *req.VersionCode); err != nil {
			return err
		}

		exists, err := appVersionExists(tx, req.Platform, *req.VersionCode, nil)
		if err != nil {
			return err
		}
		if exists {
			return ErrConflict
		}

		version = model.AppVersion{
			Platform:      req.Platform,
			VersionCode:   *req.VersionCode,
			VersionName:   req.VersionName,
			IsForceUpdate: *req.IsForceUpdate,
			ReleaseNotes:  req.ReleaseNotes,
			DownloadURL:   req.DownloadURL,
		}

		return tx.Create(&version).Error
	})
	if err != nil {
		if errors.Is(err, ErrConflict) || isDuplicateKeyError(err) {
			return model.AppVersion{}, ErrConflict
		}
		return model.AppVersion{}, err
	}

	return version, nil
}

func (s *adminAppVersionService) Update(id uuid.UUID, req dto.AdminAppVersionUpsertRequest) (model.AppVersion, error) {
	var version model.AppVersion
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&version, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		if err := lockAppVersion(tx, req.Platform, *req.VersionCode); err != nil {
			return err
		}

		exists, err := appVersionExists(tx, req.Platform, *req.VersionCode, &id)
		if err != nil {
			return err
		}
		if exists {
			return ErrConflict
		}

		version.Platform = req.Platform
		version.VersionCode = *req.VersionCode
		version.VersionName = req.VersionName
		version.IsForceUpdate = *req.IsForceUpdate
		version.ReleaseNotes = req.ReleaseNotes
		version.DownloadURL = req.DownloadURL

		return tx.Save(&version).Error
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return model.AppVersion{}, ErrNotFound
		}
		if errors.Is(err, ErrConflict) || isDuplicateKeyError(err) {
			return model.AppVersion{}, ErrConflict
		}
		return model.AppVersion{}, err
	}

	return version, nil
}

func (s *adminAppVersionService) Delete(id uuid.UUID) error {
	var version model.AppVersion
	if err := s.db.First(&version, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	return s.db.Delete(&version).Error
}

func appVersionExists(tx *gorm.DB, platform string, versionCode int, excludeID *uuid.UUID) (bool, error) {
	query := tx.Model(&model.AppVersion{}).Where("platform = ? AND version_code = ?", platform, versionCode)
	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func lockAppVersion(tx *gorm.DB, platform string, versionCode int) error {
	lockKeyPlatform := int32(crc32.ChecksumIEEE([]byte(platform)))
	lockKeyVersion := int32(versionCode)

	return tx.Exec("SELECT pg_advisory_xact_lock(?, ?)", lockKeyPlatform, lockKeyVersion).Error
}

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
