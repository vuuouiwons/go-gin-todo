package repository

import (
	"context"
	"errors"
	"strings"

	"go-gin-todo/src/db"
	"go-gin-todo/src/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TodoRepository handles database operations and reusable queries for Todo records.
type TodoRepository struct {
	db *gorm.DB
}

// NewTodoRepository constructs a new TodoRepository.
func NewTodoRepository(database *gorm.DB) *TodoRepository {
	return &TodoRepository{db: database}
}

// WithTx returns a new TodoRepository instance bound to an active transaction.
func (r *TodoRepository) WithTx(tx *gorm.DB) *TodoRepository {
	if tx == nil {
		return r
	}
	return &TodoRepository{db: tx}
}

// Create inserts a single Todo record.
func (r *TodoRepository) Create(ctx context.Context, todo *db.Todo) error {
	return r.db.WithContext(ctx).Create(todo).Error
}

// CreateBulk inserts multiple Todo records in a single batch.
func (r *TodoRepository) CreateBulk(ctx context.Context, todos *[]db.Todo) error {
	return r.db.WithContext(ctx).Create(todos).Error
}

// GetByID retrieves a single Todo by ID and userID.
func (r *TodoRepository) GetByID(ctx context.Context, id, userID uint) (*db.Todo, error) {
	var item db.Todo
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

// GetBulk retrieves paginated and filtered Todo records for a user ordered by id DESC.
func (r *TodoRepository) GetBulk(ctx context.Context, userID uint, search string, checkFilter *bool, page, limit int) ([]db.Todo, utils.Pagination, error) {
	query := r.db.WithContext(ctx).Model(&db.Todo{}).Where("user_id = ?", userID)

	search = strings.TrimSpace(search)
	if search != "" {
		query = query.Clauses(clause.Like{Column: "message", Value: "%" + search + "%"})
	}

	if checkFilter != nil {
		query = query.Where("`check` = ?", *checkFilter)
	}

	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, utils.Pagination{}, err
	}

	pagination := utils.CalculatePagination(page, limit, totalItems)
	offset := (pagination.Page - 1) * pagination.Limit
	if offset < 0 {
		offset = 0
	}

	var items []db.Todo
	if err := query.Order("id DESC").Limit(pagination.Limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, utils.Pagination{}, err
	}

	return items, pagination, nil
}

// FindByIDs retrieves multiple Todo records by user ID and a slice of IDs.
func (r *TodoRepository) FindByIDs(ctx context.Context, userID uint, ids []uint) ([]db.Todo, error) {
	var items []db.Todo
	err := r.db.WithContext(ctx).Where("user_id = ? AND id IN ?", userID, ids).Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

// Save updates an existing Todo record.
func (r *TodoRepository) Save(ctx context.Context, todo *db.Todo) error {
	return r.db.WithContext(ctx).Save(todo).Error
}

// DeleteByID removes a single Todo by ID and userID.
func (r *TodoRepository) DeleteByID(ctx context.Context, id, userID uint) (int64, error) {
	res := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&db.Todo{})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// DeleteBulk removes multiple Todo records by IDs and userID.
func (r *TodoRepository) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	res := r.db.WithContext(ctx).Where("user_id = ? AND id IN ?", userID, ids).Delete(&db.Todo{})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}
