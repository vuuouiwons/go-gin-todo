package app

import (
	"context"

	"gorm.io/gorm/clause"
)

// TodoQueries handles read operations for Todos.
// Non-reused paginated reads query directly via Queryer to avoid pass-through boilerplate.
type TodoQueries struct {
	q    Queryer
	repo TodoRepository
}

// NewTodoQueries constructs a new TodoQueries handler.
func NewTodoQueries(q Queryer, repo TodoRepository) *TodoQueries {
	return &TodoQueries{
		q:    q,
		repo: repo,
	}
}

// GetByID retrieves a single Todo owned by the user (reuses repository lookup).
func (q *TodoQueries) GetByID(ctx context.Context, id, userID uint) (*Todo, error) {
	if id == 0 {
		return nil, ErrBadRequest
	}
	return q.repo.GetByID(ctx, id, userID)
}

// GetBulk retrieves paginated Todos owned by the user with optional search and check filters.
func (q *TodoQueries) GetBulk(ctx context.Context, userID uint, search string, checkFilter *bool, page, limit int) ([]Todo, Pagination, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	query := q.q.DB(ctx).Model(&Todo{}).Where(&Todo{UserID: userID})

	if search != "" {
		query = query.Clauses(clause.Like{Column: "message", Value: "%" + search + "%"})
	}

	if checkFilter != nil {
		query = query.Where(map[string]interface{}{"check": *checkFilter})
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, Pagination{}, err
	}

	offset := (page - 1) * limit
	var items []Todo
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}, Desc: true}).
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return nil, Pagination{}, err
	}

	return items, NewPagination(page, limit, total), nil
}
