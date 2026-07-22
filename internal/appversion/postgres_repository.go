package appversion

import (
	"errors"
	"hash/crc32"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type PostgresRepository struct{ db *gorm.DB }

func NewPostgresRepository(db *gorm.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) List() ([]Entity, error) {
	var entities []Entity
	return entities, r.db.Order("platform asc, version_code desc").Find(&entities).Error
}

func (r *PostgresRepository) ListByPlatform(platform string) ([]Entity, error) {
	var entities []Entity
	return entities, r.db.Where("platform = ?", platform).Order("version_code desc").Find(&entities).Error
}

func (r *PostgresRepository) Get(id uuid.UUID) (Entity, error) {
	var entity Entity
	if err := r.db.First(&entity, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Entity{}, ErrNotFound
		}
		return Entity{}, err
	}
	return entity, nil
}

func (r *PostgresRepository) LatestByPlatform(platform string) (*Entity, error) {
	var entity Entity
	if err := r.db.Where("platform = ?", platform).Order("version_code desc").First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

func (r *PostgresRepository) HasForceUpdateAfter(platform string, versionCode int) (bool, error) {
	var entity Entity
	err := r.db.Select("id").Where("platform = ? AND version_code > ? AND is_force_update = ?", platform, versionCode, true).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (r *PostgresRepository) Create(entity Entity) (Entity, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := lock(tx, entity.Platform, entity.VersionCode); err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&Entity{}).Where("platform = ? AND version_code = ?", entity.Platform, entity.VersionCode).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrConflict
		}
		return tx.Create(&entity).Error
	})
	if errors.Is(err, ErrConflict) || isDuplicateKey(err) {
		return Entity{}, ErrConflict
	}
	return entity, err
}

func (r *PostgresRepository) Update(entity Entity) (Entity, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := lock(tx, entity.Platform, entity.VersionCode); err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&Entity{}).Where("platform = ? AND version_code = ? AND id <> ?", entity.Platform, entity.VersionCode, entity.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrConflict
		}
		return tx.Save(&entity).Error
	})
	if errors.Is(err, ErrConflict) || isDuplicateKey(err) {
		return Entity{}, ErrConflict
	}
	return entity, err
}

func (r *PostgresRepository) Delete(id uuid.UUID) error {
	entity, err := r.Get(id)
	if err != nil {
		return err
	}
	return r.db.Delete(&entity).Error
}

func lock(tx *gorm.DB, platform string, versionCode int) error {
	return tx.Exec("SELECT pg_advisory_xact_lock(?, ?)", int32(crc32.ChecksumIEEE([]byte(platform))), int32(versionCode)).Error
}

func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
