package data

import (
	"context"
	"errors"
	"strings"

	"go-gin-todo/src/business"
	"go-gin-todo/src/shared"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TodolistRepository implements business.ITodolistRepository using GORM.
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
func (r *TodolistRepository) Create(ctx context.Context, item *business.Todolist) error {
	entity := TodolistToEntity(item)
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return err
	}
	item.ID = entity.ID
	item.CreatedAt = entity.CreatedAt
	item.UpdatedAt = entity.UpdatedAt
	return nil
}

// CreateBulk inserts multiple Todolist records in a single batch.
func (r *TodolistRepository) CreateBulk(ctx context.Context, items *[]business.Todolist) error {
	entities := make([]TodolistEntity, len(*items))
	for i, item := range *items {
		entities[i] = *TodolistToEntity(&item)
	}

	if err := r.db.WithContext(ctx).Create(&entities).Error; err != nil {
		return err
	}

	for i := range *items {
		(*items)[i].ID = entities[i].ID
		(*items)[i].CreatedAt = entities[i].CreatedAt
		(*items)[i].UpdatedAt = entities[i].UpdatedAt
	}

	return nil
}

// GetByID retrieves a single Todolist by ID and userID.
func (r *TodolistRepository) GetByID(ctx context.Context, id, userID uint) (*business.Todolist, error) {
	var entity TodolistEntity
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return TodolistToDomain(&entity), nil
}

// GetBulk retrieves paginated and filtered Todolist records for a user ordered by id DESC.
func (r *TodolistRepository) GetBulk(ctx context.Context, userID uint, search string, page, limit int) ([]business.Todolist, shared.Pagination, error) {
	query := r.db.WithContext(ctx).Model(&TodolistEntity{}).Where("user_id = ?", userID)

	search = strings.TrimSpace(search)
	if search != "" {
		query = query.Clauses(clause.Like{Column: "title", Value: "%" + search + "%"})
	}

	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, shared.Pagination{}, err
	}

	pagination := shared.CalculatePagination(page, limit, totalItems)
	offset := (pagination.Page - 1) * pagination.Limit
	if offset < 0 {
		offset = 0
	}

	var entities []TodolistEntity
	if err := query.Order("id DESC").Limit(pagination.Limit).Offset(offset).Find(&entities).Error; err != nil {
		return nil, shared.Pagination{}, err
	}

	return TodolistListToDomain(entities), pagination, nil
}

// FindByIDs retrieves multiple Todolist records by user ID and a slice of IDs.
func (r *TodolistRepository) FindByIDs(ctx context.Context, userID uint, ids []uint) ([]business.Todolist, error) {
	var entities []TodolistEntity
	err := r.db.WithContext(ctx).Where("user_id = ? AND id IN ?", userID, ids).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return TodolistListToDomain(entities), nil
}

// Save updates an existing Todolist record.
func (r *TodolistRepository) Save(ctx context.Context, item *business.Todolist) error {
	entity := TodolistToEntity(item)
	return r.db.WithContext(ctx).Save(entity).Error
}

// UpdateBulk atomically updates multiple Todolists owned by the user using a GORM database transaction.
func (r *TodolistRepository) UpdateBulk(ctx context.Context, userID uint, updates []business.Todolist) ([]business.Todolist, error) {
	ids := make([]uint, len(updates))
	for i, u := range updates {
		ids[i] = u.ID
	}

	var result []business.Todolist

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var entities []TodolistEntity
		if err := tx.Where("user_id = ? AND id IN ?", userID, ids).Find(&entities).Error; err != nil {
			return shared.ErrInternalServer
		}

		existingMap := make(map[uint]TodolistEntity, len(entities))
		for _, entity := range entities {
			existingMap[entity.ID] = entity
		}

		for _, u := range updates {
			if _, found := existingMap[u.ID]; !found {
				return shared.ErrNotFound
			}
		}

		result = make([]business.Todolist, len(updates))
		for i, u := range updates {
			entity := existingMap[u.ID]
			entity.Title = u.Title
			if err := tx.Save(&entity).Error; err != nil {
				return shared.ErrInternalServer
			}
			result[i] = *TodolistToDomain(&entity)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteByID removes a single Todolist by ID and userID.
func (r *TodolistRepository) DeleteByID(ctx context.Context, id, userID uint) (int64, error) {
	res := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&TodolistEntity{})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// DeleteBulk removes multiple Todolist records by IDs and userID.
func (r *TodolistRepository) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	res := r.db.WithContext(ctx).Where("user_id = ? AND id IN ?", userID, ids).Delete(&TodolistEntity{})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}
