package app

import (
	"context"
)

// TodolistQueries handles read operations for Todolists.
// Non-reused paginated reads query directly via Queryer to avoid pass-through boilerplate.
type TodolistQueries struct {
	q    Queryer
	repo TodolistRepository
}

// NewTodolistQueries constructs a new TodolistQueries handler.
func NewTodolistQueries(q Queryer, repo TodolistRepository) *TodolistQueries {
	return &TodolistQueries{
		q:    q,
		repo: repo,
	}
}

// GetByID retrieves a single Todolist owned by the user (reuses repository lookup).
func (q *TodolistQueries) GetByID(ctx context.Context, id, userID uint) (*Todolist, error) {
	if id == 0 {
		return nil, ErrBadRequest
	}
	return q.repo.GetByID(ctx, id, userID)
}

// GetBulk retrieves paginated Todolists owned by the user with optional title search.
func (q *TodolistQueries) GetBulk(ctx context.Context, userID uint, search string, page, limit int) ([]Todolist, Pagination, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	base := q.q.DB(ctx).Model(&Todolist{}).Where("user_id = ?", userID)
	if search != "" {
		base = base.Where("title LIKE ?", "%"+search+"%")
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, Pagination{}, err
	}

	offset := (page - 1) * limit
	var items []Todolist
	if err := base.Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, Pagination{}, err
	}

	return items, NewPagination(page, limit, total), nil
}
