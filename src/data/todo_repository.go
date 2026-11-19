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

// TodoRepository implements business.ITodoRepository using GORM.
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
func (r *TodoRepository) Create(ctx context.Context, todo *business.Todo) error {
	entity := TodoToEntity(todo)
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return err
	}
	todo.ID = entity.ID
	todo.CreatedAt = entity.CreatedAt
	todo.UpdatedAt = entity.UpdatedAt
	return nil
}

// CreateBulk inserts multiple Todo records in a single batch.
func (r *TodoRepository) CreateBulk(ctx context.Context, todos *[]business.Todo) error {
	entities := make([]TodoEntity, len(*todos))
	for i, todo := range *todos {
		entities[i] = *TodoToEntity(&todo)
	}

	if err := r.db.WithContext(ctx).Create(&entities).Error; err != nil {
		return err
	}

	for i := range *todos {
		(*todos)[i].ID = entities[i].ID
		(*todos)[i].CreatedAt = entities[i].CreatedAt
		(*todos)[i].UpdatedAt = entities[i].UpdatedAt
	}

	return nil
}

// GetByID retrieves a single Todo by ID and userID.
func (r *TodoRepository) GetByID(ctx context.Context, id, userID uint) (*business.Todo, error) {
	var entity TodoEntity
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return TodoToDomain(&entity), nil
}

// GetBulk retrieves paginated and filtered Todo records for a user ordered by id DESC.
func (r *TodoRepository) GetBulk(ctx context.Context, userID uint, search string, checkFilter *bool, page, limit int) ([]business.Todo, shared.Pagination, error) {
	query := r.db.WithContext(ctx).Model(&TodoEntity{}).Where("user_id = ?", userID)

	search = strings.TrimSpace(search)
	if search != "" {
		query = query.Clauses(clause.Like{Column: "message", Value: "%" + search + "%"})
	}

	if checkFilter != nil {
		query = query.Where("`check` = ?", *checkFilter)
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

	var entities []TodoEntity
	if err := query.Order("id DESC").Limit(pagination.Limit).Offset(offset).Find(&entities).Error; err != nil {
		return nil, shared.Pagination{}, err
	}

	return TodoListToDomain(entities), pagination, nil
}

// FindByIDs retrieves multiple Todo records by user ID and a slice of IDs.
func (r *TodoRepository) FindByIDs(ctx context.Context, userID uint, ids []uint) ([]business.Todo, error) {
	var entities []TodoEntity
	err := r.db.WithContext(ctx).Where("user_id = ? AND id IN ?", userID, ids).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return TodoListToDomain(entities), nil
}

// Save updates an existing Todo record.
func (r *TodoRepository) Save(ctx context.Context, todo *business.Todo) error {
	entity := TodoToEntity(todo)
	return r.db.WithContext(ctx).Save(entity).Error
}

// UpdateBulk atomically updates multiple Todos owned by the user using a GORM database transaction.
func (r *TodoRepository) UpdateBulk(ctx context.Context, userID uint, updates []business.TodoUpdateItem) ([]business.Todo, error) {
	ids := make([]uint, len(updates))
	for i, item := range updates {
		ids[i] = item.ID
	}

	var result []business.Todo

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var entities []TodoEntity
		if err := tx.Where("user_id = ? AND id IN ?", userID, ids).Find(&entities).Error; err != nil {
			return shared.ErrInternalServer
		}

		existingMap := make(map[uint]TodoEntity, len(entities))
		for _, entity := range entities {
			existingMap[entity.ID] = entity
		}

		for _, item := range updates {
			if _, found := existingMap[item.ID]; !found {
				return shared.ErrNotFound
			}
		}

		result = make([]business.Todo, len(updates))
		for i, item := range updates {
			entity := existingMap[item.ID]
			if item.Message != nil {
				entity.Message = *item.Message
			}
			if item.Check != nil {
				entity.Check = *item.Check
			}
			if err := tx.Save(&entity).Error; err != nil {
				return shared.ErrInternalServer
			}
			result[i] = *TodoToDomain(&entity)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteByID removes a single Todo by ID and userID.
func (r *TodoRepository) DeleteByID(ctx context.Context, id, userID uint) (int64, error) {
	res := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&TodoEntity{})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// DeleteBulk removes multiple Todo records by IDs and userID.
func (r *TodoRepository) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	res := r.db.WithContext(ctx).Where("user_id = ? AND id IN ?", userID, ids).Delete(&TodoEntity{})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}
