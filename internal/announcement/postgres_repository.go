package announcement

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresRepository struct{ db *gorm.DB }

func NewPostgresRepository(db *gorm.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) List(ctx context.Context) ([]Entity, error) {
	var entities []Entity
	return entities, r.db.WithContext(ctx).Order("created_at desc").Find(&entities).Error
}

func (r *PostgresRepository) ListActive(ctx context.Context, platform string) ([]Entity, error) {
	var entities []Entity
	return entities, r.db.WithContext(ctx).
		Where("is_active = ? AND platform IN ?", true, []string{platform, PlatformAll}).
		Order("created_at desc").
		Find(&entities).Error
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

func (r *PostgresRepository) Create(ctx context.Context, entity Entity) (Entity, error) {
	err := r.db.WithContext(ctx).Model(&Entity{}).Create(map[string]any{
		"id":         entity.ID,
		"title":      entity.Title,
		"content":    entity.Content,
		"platform":   entity.Platform,
		"is_active":  entity.IsActive,
		"is_banner":  entity.IsBanner,
		"created_at": entity.CreatedAt,
	}).Error
	return entity, err
}
func (r *PostgresRepository) Update(ctx context.Context, id uuid.UUID, values Entity) (Entity, error) {
	var entity Entity
	result := r.db.WithContext(ctx).
		Model(&entity).
		Clauses(clause.Returning{}).
		Where("id = ?", id).
		Select("title", "content", "platform", "is_active", "is_banner").
		Updates(values)
	if result.Error != nil {
		return Entity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return Entity{}, ErrNotFound
	}
	return entity, nil
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
