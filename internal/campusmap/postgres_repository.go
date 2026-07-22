package campusmap

import (
	"context"

	"gorm.io/gorm"
)

type PostgresRepository struct{ db *gorm.DB }

func NewPostgresRepository(db *gorm.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) List(ctx context.Context) ([]Entity, error) {
	var entities []Entity
	return entities, r.db.WithContext(ctx).Find(&entities).Error
}
