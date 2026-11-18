package infra

import (
	"context"
	"errors"

	"go-gin-todo/internal/app"
	"gorm.io/gorm"
)

type todolistRepository struct {
	db *gorm.DB
}

// NewTodolistRepository constructs a GORM implementation of app.TodolistRepository.
func NewTodolistRepository(db *gorm.DB) app.TodolistRepository {
	return &todolistRepository{db: db}
}

func (r *todolistRepository) Create(ctx context.Context, todolist *app.Todolist) error {
	return dbFromContext(ctx, r.db).Create(todolist).Error
}

func (r *todolistRepository) CreateBulk(ctx context.Context, todolists []app.Todolist) ([]app.Todolist, error) {
	if err := dbFromContext(ctx, r.db).Create(&todolists).Error; err != nil {
		return nil, err
	}
	return todolists, nil
}

func (r *todolistRepository) GetByID(ctx context.Context, id, userID uint) (*app.Todolist, error) {
	var item app.Todolist
	if err := dbFromContext(ctx, r.db).Where("id = ? AND user_id = ?", id, userID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, app.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *todolistRepository) Update(ctx context.Context, todolist *app.Todolist) error {
	return dbFromContext(ctx, r.db).Save(todolist).Error
}

func (r *todolistRepository) UpdateBulk(ctx context.Context, userID uint, updates []app.Todolist) ([]app.Todolist, error) {
	if len(updates) == 0 {
		return nil, nil
	}

	ids := make([]uint, len(updates))
	for i, u := range updates {
		ids[i] = u.ID
	}

	var existing []app.Todolist
	if err := dbFromContext(ctx, r.db).Where("user_id = ? AND id IN ?", userID, ids).Find(&existing).Error; err != nil {
		return nil, err
	}

	existingMap := make(map[uint]app.Todolist, len(existing))
	for _, item := range existing {
		existingMap[item.ID] = item
	}

	for _, u := range updates {
		if _, found := existingMap[u.ID]; !found {
			return nil, app.ErrNotFound
		}
	}

	result := make([]app.Todolist, len(updates))
	for i, u := range updates {
		item := existingMap[u.ID]
		item.Title = u.Title
		if err := dbFromContext(ctx, r.db).Save(&item).Error; err != nil {
			return nil, err
		}
		result[i] = item
	}

	return result, nil
}

func (r *todolistRepository) DeleteByID(ctx context.Context, id, userID uint) (int64, error) {
	result := dbFromContext(ctx, r.db).Where("id = ? AND user_id = ?", id, userID).Delete(&app.Todolist{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (r *todolistRepository) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	result := dbFromContext(ctx, r.db).Where("user_id = ? AND id IN ?", userID, ids).Delete(&app.Todolist{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
