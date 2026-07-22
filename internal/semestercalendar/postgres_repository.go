package semestercalendar

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type PostgresRepository struct{ db *gorm.DB }

func NewPostgresRepository(db *gorm.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) List() ([]Entity, error) {
	var entities []Entity
	return entities, r.db.Order("semester_code desc").Find(&entities).Error
}

func (r *PostgresRepository) ListSummaries() ([]Entity, error) {
	var entities []Entity
	return entities, r.db.Select("semester_code", "title", "subtitle").Order("semester_code desc").Find(&entities).Error
}

func (r *PostgresRepository) Get(code string) (Entity, error) {
	var entity Entity
	if err := r.db.Where("semester_code = ?", code).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Entity{}, ErrNotFound
		}
		return Entity{}, err
	}
	return entity, nil
}

func (r *PostgresRepository) Create(entity Entity) (Entity, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&entity).Error
	})
	if isDuplicateKey(err) {
		return Entity{}, ErrConflict
	}
	return entity, err
}

func (r *PostgresRepository) Update(entity Entity) (Entity, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Save(&entity).Error
	})
	if isDuplicateKey(err) {
		return Entity{}, ErrConflict
	}
	return entity, err
}

func (r *PostgresRepository) Delete(code string) error {
	entity, err := r.Get(code)
	if err != nil {
		return err
	}
	return r.db.Delete(&entity).Error
}

func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
