package repository

import (
	"context"
	"errors"

	"go-gin-todo/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type todoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) domain.TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) Create(ctx context.Context, todo *domain.Todo) error {
	return r.db.WithContext(ctx).Create(todo).Error
}

func (r *todoRepository) CreateBulk(ctx context.Context, todos []domain.Todo) ([]domain.Todo, error) {
	if err := r.db.WithContext(ctx).Create(&todos).Error; err != nil {
		return nil, err
	}
	return todos, nil
}

func (r *todoRepository) GetByID(ctx context.Context, id uint, userID uint) (*domain.Todo, error) {
	var item domain.Todo
	if err := r.db.WithContext(ctx).Where(&domain.Todo{ID: id, UserID: userID}).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *todoRepository) GetBulk(ctx context.Context, userID uint, search string, checkFilter *bool, page, pageSize int) ([]domain.Todo, int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.Todo{}).Where(&domain.Todo{UserID: userID})

	if search != "" {
		query = query.Clauses(clause.Like{Column: "message", Value: "%" + search + "%"})
	}

	if checkFilter != nil {
		query = query.Where(map[string]interface{}{"check": *checkFilter})
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	var items []domain.Todo
	if err := query.Order("id DESC").Limit(pageSize).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *todoRepository) Update(ctx context.Context, todo *domain.Todo) error {
	return r.db.WithContext(ctx).Save(todo).Error
}

func (r *todoRepository) UpdateBulk(ctx context.Context, userID uint, updates []domain.TodoUpdateItem) ([]domain.Todo, error) {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var updated []domain.Todo
	for _, item := range updates {
		var existing domain.Todo
		if err := tx.Where(&domain.Todo{ID: item.ID, UserID: userID}).First(&existing).Error; err != nil {
			tx.Rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, domain.ErrNotFound
			}
			return nil, err
		}

		if item.Message != nil {
			existing.Message = *item.Message
		}
		if item.Check != nil {
			existing.Check = *item.Check
		}

		if err := tx.Save(&existing).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		updated = append(updated, existing)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return updated, nil
}

func (r *todoRepository) DeleteByID(ctx context.Context, id uint, userID uint) (int64, error) {
	result := r.db.WithContext(ctx).Where(&domain.Todo{ID: id, UserID: userID}).Delete(&domain.Todo{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (r *todoRepository) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	result := r.db.WithContext(ctx).Where(&domain.Todo{UserID: userID}).Where(map[string]interface{}{"id": ids}).Delete(&domain.Todo{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
