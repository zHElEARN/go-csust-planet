package appversion

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresRepository struct{ db *gorm.DB }

func NewPostgresRepository(db *gorm.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) List(ctx context.Context) ([]Entity, error) {
	var entities []Entity
	return entities, r.db.WithContext(ctx).Order("platform asc, version_code desc").Find(&entities).Error
}

func (r *PostgresRepository) ListByPlatform(ctx context.Context, platform string) ([]Entity, error) {
	var entities []Entity
	return entities, r.db.WithContext(ctx).Where("platform = ?", platform).Order("version_code desc").Find(&entities).Error
}

func (r *PostgresRepository) Get(ctx context.Context, id uuid.UUID) (Entity, error) {
	var entity Entity
	if err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Entity{}, ErrNotFound
		}
		return Entity{}, err
	}
	return entity, nil
}

func (r *PostgresRepository) LatestByPlatform(ctx context.Context, platform string) (*Entity, error) {
	var entity Entity
	if err := r.db.WithContext(ctx).Where("platform = ?", platform).Order("version_code desc").First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

func (r *PostgresRepository) HasForceUpdateAfter(ctx context.Context, platform string, versionCode int) (bool, error) {
	var entity Entity
	err := r.db.WithContext(ctx).Select("id").Where("platform = ? AND version_code > ? AND is_force_update = ?", platform, versionCode, true).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (r *PostgresRepository) Create(ctx context.Context, entity Entity) (Entity, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Create(&entity).Error
	})
	if isDuplicateKey(err) {
		return Entity{}, ErrConflict
	}
	return entity, err
}

func (r *PostgresRepository) Update(ctx context.Context, id uuid.UUID, values Entity) (Entity, error) {
	var entity Entity
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&entity).
			Clauses(clause.Returning{}).
			Where("id = ?", id).
			Select("platform", "version_code", "version_name", "is_force_update", "release_notes", "download_url").
			Updates(values)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
	if isDuplicateKey(err) {
		return Entity{}, ErrConflict
	}
	return entity, err
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&Entity{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
