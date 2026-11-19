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

// TodolistRepository handles database operations and reusable queries for Todolist records.
type TodolistRepository struct {
	db *gorm.DB
}

// NewTodolistRepository constructs a new TodolistRepository.
func NewTodolistRepository(database *gorm.DB) *TodolistRepository {
	return &TodolistRepository{db: database}
}

// WithTx returns a new TodolistRepository instance bound to an active transaction.
func (r *TodolistRepository) WithTx(tx *gorm.DB) *TodolistRepository {
	if tx == nil {
		return r
	}
	return &TodolistRepository{db: tx}
}

// Create inserts a single Todolist record.
func (r *TodolistRepository) Create(ctx context.Context, item *db.Todolist) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// CreateBulk inserts multiple Todolist records in a single batch.
func (r *TodolistRepository) CreateBulk(ctx context.Context, items *[]db.Todolist) error {
	return r.db.WithContext(ctx).Create(items).Error
}

// GetByID retrieves a single Todolist by ID and userID.
func (r *TodolistRepository) GetByID(ctx context.Context, id, userID uint) (*db.Todolist, error) {
	var item db.Todolist
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

// GetBulk retrieves paginated and filtered Todolist records for a user ordered by id DESC.
func (r *TodolistRepository) GetBulk(ctx context.Context, userID uint, search string, page, limit int) ([]db.Todolist, utils.Pagination, error) {
	query := r.db.WithContext(ctx).Model(&db.Todolist{}).Where("user_id = ?", userID)

	search = strings.TrimSpace(search)
	if search != "" {
		query = query.Clauses(clause.Like{Column: "title", Value: "%" + search + "%"})
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

	var items []db.Todolist
	if err := query.Order("id DESC").Limit(pagination.Limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, utils.Pagination{}, err
	}

	return items, pagination, nil
}

// FindByIDs retrieves multiple Todolist records by user ID and a slice of IDs.
func (r *TodolistRepository) FindByIDs(ctx context.Context, userID uint, ids []uint) ([]db.Todolist, error) {
	var items []db.Todolist
	err := r.db.WithContext(ctx).Where("user_id = ? AND id IN ?", userID, ids).Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

// Save updates an existing Todolist record.
func (r *TodolistRepository) Save(ctx context.Context, item *db.Todolist) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// DeleteByID removes a single Todolist by ID and userID.
func (r *TodolistRepository) DeleteByID(ctx context.Context, id, userID uint) (int64, error) {
	res := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&db.Todolist{})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// DeleteBulk removes multiple Todolist records by IDs and userID.
func (r *TodolistRepository) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	res := r.db.WithContext(ctx).Where("user_id = ? AND id IN ?", userID, ids).Delete(&db.Todolist{})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}
