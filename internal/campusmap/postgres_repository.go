package campusmap

import "gorm.io/gorm"

type PostgresRepository struct{ db *gorm.DB }

func NewPostgresRepository(db *gorm.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) List() ([]Entity, error) {
	var entities []Entity
	return entities, r.db.Find(&entities).Error
}
