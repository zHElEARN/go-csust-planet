package semestercalendar

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresRepository struct{ db *gorm.DB }

func NewPostgresRepository(db *gorm.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) List(ctx context.Context) ([]Entity, error) {
	var entities []Entity
	return entities, r.db.WithContext(ctx).Order("semester_code desc").Find(&entities).Error
}

func (r *PostgresRepository) ListSummaries(ctx context.Context) ([]Entity, error) {
	var entities []Entity
	return entities, r.db.WithContext(ctx).Select("semester_code", "title", "subtitle").Order("semester_code desc").Find(&entities).Error
}

func (r *PostgresRepository) Get(ctx context.Context, code string) (Entity, error) {
	var entity Entity
	if err := r.db.WithContext(ctx).Where("semester_code = ?", code).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Entity{}, ErrNotFound
		}
		return Entity{}, err
	}
	return entity, nil
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

func (r *PostgresRepository) Update(ctx context.Context, code string, values Entity) (Entity, error) {
	var entity Entity
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&entity).
			Clauses(clause.Returning{}).
			Where("semester_code = ?", code).
			Select(
				"semester_code", "title", "subtitle",
				"calendar_start", "calendar_end", "semester_start", "semester_end",
				"notes", "custom_week_ranges",
			).
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

func (r *PostgresRepository) Delete(ctx context.Context, code string) error {
	result := r.db.WithContext(ctx).Where("semester_code = ?", code).Delete(&Entity{})
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
