package announcement

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostgresRepository struct{ db *gorm.DB }

func NewPostgresRepository(db *gorm.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) List() ([]Entity, error) {
	var entities []Entity
	return entities, r.db.Order("created_at desc").Find(&entities).Error
}

func (r *PostgresRepository) ListActive() ([]Entity, error) {
	var entities []Entity
	return entities, r.db.Where("is_active = ?", true).Order("created_at desc").Find(&entities).Error
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

func (r *PostgresRepository) Create(entity Entity) (Entity, error) {
	err := r.db.Model(&Entity{}).Create(map[string]any{
		"id":         entity.ID,
		"title":      entity.Title,
		"content":    entity.Content,
		"is_active":  entity.IsActive,
		"is_banner":  entity.IsBanner,
		"created_at": entity.CreatedAt,
	}).Error
	return entity, err
}
func (r *PostgresRepository) Update(entity Entity) (Entity, error) {
	return entity, r.db.Save(&entity).Error
}

func (r *PostgresRepository) Delete(id uuid.UUID) error {
	entity, err := r.Get(id)
	if err != nil {
		return err
	}
	return r.db.Delete(&entity).Error
}
