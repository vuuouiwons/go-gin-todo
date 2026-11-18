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
	if err := dbFromContext(ctx, r.db).Where(&app.Todo{ID: id, UserID: userID}).First(&item).Error; err != nil {
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

func (r *todoRepository) DeleteByID(ctx context.Context, id, userID uint) (int64, error) {
	result := dbFromContext(ctx, r.db).Where(&app.Todo{ID: id, UserID: userID}).Delete(&app.Todo{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (r *todoRepository) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	result := dbFromContext(ctx, r.db).Where(&app.Todo{UserID: userID}).Where(map[string]interface{}{"id": ids}).Delete(&app.Todo{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
