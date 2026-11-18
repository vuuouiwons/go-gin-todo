package infra

import (
	"context"
	"errors"

	"go-gin-todo/internal/app"
	"gorm.io/gorm"
)

type todoRepository struct {
	db *gorm.DB
}

// NewTodoRepository constructs a GORM implementation of app.TodoRepository.
func NewTodoRepository(db *gorm.DB) app.TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) Create(ctx context.Context, todo *app.Todo) error {
	return dbFromContext(ctx, r.db).Create(todo).Error
}

func (r *todoRepository) CreateBulk(ctx context.Context, todos []app.Todo) ([]app.Todo, error) {
	if err := dbFromContext(ctx, r.db).Create(&todos).Error; err != nil {
		return nil, err
	}
	return todos, nil
}

func (r *todoRepository) GetByID(ctx context.Context, id, userID uint) (*app.Todo, error) {
	var item app.Todo
	if err := dbFromContext(ctx, r.db).Where("id = ? AND user_id = ?", id, userID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, app.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *todoRepository) Update(ctx context.Context, todo *app.Todo) error {
	return dbFromContext(ctx, r.db).Save(todo).Error
}

func (r *todoRepository) UpdateBulk(ctx context.Context, userID uint, updates []app.TodoUpdateItem) ([]app.Todo, error) {
	if len(updates) == 0 {
		return nil, nil
	}

	ids := make([]uint, len(updates))
	for i, u := range updates {
		ids[i] = u.ID
	}

	var existing []app.Todo
	if err := dbFromContext(ctx, r.db).Where("user_id = ? AND id IN ?", userID, ids).Find(&existing).Error; err != nil {
		return nil, err
	}

	existingMap := make(map[uint]app.Todo, len(existing))
	for _, item := range existing {
		existingMap[item.ID] = item
	}

	for _, u := range updates {
		if _, found := existingMap[u.ID]; !found {
			return nil, app.ErrNotFound
		}
	}

	result := make([]app.Todo, len(updates))
	for i, u := range updates {
		item := existingMap[u.ID]
		if u.Message != nil {
			item.Message = *u.Message
		}
		if u.Check != nil {
			item.Check = *u.Check
		}
		if err := dbFromContext(ctx, r.db).Save(&item).Error; err != nil {
			return nil, err
		}
		result[i] = item
	}

	return result, nil
}

func (r *todoRepository) DeleteByID(ctx context.Context, id, userID uint) (int64, error) {
	result := dbFromContext(ctx, r.db).Where("id = ? AND user_id = ?", id, userID).Delete(&app.Todo{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (r *todoRepository) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	result := dbFromContext(ctx, r.db).Where("user_id = ? AND id IN ?", userID, ids).Delete(&app.Todo{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
