package repository

import (
	"context"
	"errors"

	"go-gin-todo/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type todolistRepository struct {
	db *gorm.DB
}

func NewTodolistRepository(db *gorm.DB) domain.TodolistRepository {
	return &todolistRepository{db: db}
}

func (r *todolistRepository) Create(ctx context.Context, todolist *domain.Todolist) error {
	return r.db.WithContext(ctx).Create(todolist).Error
}

func (r *todolistRepository) CreateBulk(ctx context.Context, todolists []domain.Todolist) ([]domain.Todolist, error) {
	if err := r.db.WithContext(ctx).Create(&todolists).Error; err != nil {
		return nil, err
	}
	return todolists, nil
}

func (r *todolistRepository) GetByID(ctx context.Context, id uint, userID uint) (*domain.Todolist, error) {
	var item domain.Todolist
	if err := r.db.WithContext(ctx).Where(&domain.Todolist{ID: id, UserID: userID}).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *todolistRepository) GetBulk(ctx context.Context, userID uint, search string, page, limit int) ([]domain.Todolist, int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.Todolist{}).Where(&domain.Todolist{UserID: userID})

	if search != "" {
		query = query.Clauses(clause.Like{Column: "title", Value: "%" + search + "%"})
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	var items []domain.Todolist
	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *todolistRepository) Update(ctx context.Context, todolist *domain.Todolist) error {
	return r.db.WithContext(ctx).Save(todolist).Error
}

func (r *todolistRepository) UpdateBulk(ctx context.Context, userID uint, updates []domain.Todolist) ([]domain.Todolist, error) {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var updated []domain.Todolist
	for _, item := range updates {
		var existing domain.Todolist
		if err := tx.Where(&domain.Todolist{ID: item.ID, UserID: userID}).First(&existing).Error; err != nil {
			tx.Rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, domain.ErrNotFound
			}
			return nil, err
		}

		existing.Title = item.Title
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

func (r *todolistRepository) DeleteByID(ctx context.Context, id uint, userID uint) (int64, error) {
	result := r.db.WithContext(ctx).Where(&domain.Todolist{ID: id, UserID: userID}).Delete(&domain.Todolist{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (r *todolistRepository) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	result := r.db.WithContext(ctx).Where(&domain.Todolist{UserID: userID}).Where(map[string]interface{}{"id": ids}).Delete(&domain.Todolist{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
