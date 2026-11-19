package services

import (
	"context"
	"strings"

	"go-gin-todo/src/db"
	"go-gin-todo/src/db/repository"
	"go-gin-todo/src/utils"
	"gorm.io/gorm"
)

// TodoService encapsulates business logic and coordinates repository operations.
type TodoService struct {
	db   *gorm.DB
	repo *repository.TodoRepository
}

// NewTodoService constructs a new TodoService with a reusable TodoRepository and DB handle for transactions.
func NewTodoService(database *gorm.DB, repo *repository.TodoRepository) *TodoService {
	return &TodoService{
		db:   database,
		repo: repo,
	}
}

// CreateOne validates and creates a single Todo item for a user.
func (s *TodoService) CreateOne(ctx context.Context, userID uint, message string, check bool) (*db.Todo, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, utils.ErrBadRequest
	}

	todo := &db.Todo{
		UserID:  userID,
		Message: message,
		Check:   check,
	}

	if err := s.repo.Create(ctx, todo); err != nil {
		return nil, utils.ErrInternalServer
	}

	return todo, nil
}

// CreateBulk validates and creates multiple Todo items for a user.
func (s *TodoService) CreateBulk(ctx context.Context, userID uint, items []db.Todo) ([]db.Todo, error) {
	if len(items) == 0 {
		return nil, utils.ErrBadRequest
	}

	todos := make([]db.Todo, 0, len(items))
	for _, item := range items {
		msg := strings.TrimSpace(item.Message)
		if msg == "" {
			return nil, utils.ErrBadRequest
		}
		todos = append(todos, db.Todo{
			UserID:  userID,
			Message: msg,
			Check:   item.Check,
		})
	}

	if err := s.repo.CreateBulk(ctx, &todos); err != nil {
		return nil, utils.ErrInternalServer
	}

	return todos, nil
}

// GetByID retrieves a single Todo by ID verifying ownership.
func (s *TodoService) GetByID(ctx context.Context, id, userID uint) (*db.Todo, error) {
	if id == 0 {
		return nil, utils.ErrBadRequest
	}

	item, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	return item, nil
}

// GetBulk retrieves paginated and filtered Todos for a user ordered by id DESC.
func (s *TodoService) GetBulk(ctx context.Context, userID uint, search string, checkFilter *bool, page, limit int) ([]db.Todo, db.Pagination, error) {
	items, pagination, err := s.repo.GetBulk(ctx, userID, search, checkFilter, page, limit)
	if err != nil {
		return nil, db.Pagination{}, utils.ErrInternalServer
	}

	return items, pagination, nil
}

// UpdateByID updates message and/or check status of a user-owned Todo.
func (s *TodoService) UpdateByID(ctx context.Context, id, userID uint, message *string, check *bool) (*db.Todo, error) {
	if id == 0 || (message == nil && check == nil) {
		return nil, utils.ErrBadRequest
	}

	var trimmedMsg *string
	if message != nil {
		msg := strings.TrimSpace(*message)
		if msg == "" {
			return nil, utils.ErrBadRequest
		}
		trimmedMsg = &msg
	}

	item, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if trimmedMsg != nil {
		item.Message = *trimmedMsg
	}
	if check != nil {
		item.Check = *check
	}

	if err := s.repo.Save(ctx, item); err != nil {
		return nil, utils.ErrInternalServer
	}

	return item, nil
}

// UpdateBulk atomically updates multiple Todos owned by the user using a GORM database transaction.
func (s *TodoService) UpdateBulk(ctx context.Context, userID uint, updates []db.TodoUpdateItem) ([]db.Todo, error) {
	if len(updates) == 0 {
		return nil, utils.ErrBadRequest
	}

	cleanUpdates := make([]db.TodoUpdateItem, len(updates))
	ids := make([]uint, len(updates))
	for i, item := range updates {
		if item.ID == 0 {
			return nil, utils.ErrBadRequest
		}
		var cleanMsg *string
		if item.Message != nil {
			msg := strings.TrimSpace(*item.Message)
			if msg == "" {
				return nil, utils.ErrBadRequest
			}
			cleanMsg = &msg
		}
		cleanUpdates[i] = db.TodoUpdateItem{
			ID:      item.ID,
			Message: cleanMsg,
			Check:   item.Check,
		}
		ids[i] = item.ID
	}

	var result []db.Todo

	// Typical GORM transaction pattern with repository scoping via WithTx
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		existing, err := txRepo.FindByIDs(ctx, userID, ids)
		if err != nil {
			return utils.ErrInternalServer
		}

		existingMap := make(map[uint]db.Todo, len(existing))
		for _, item := range existing {
			existingMap[item.ID] = item
		}

		for _, item := range cleanUpdates {
			if _, found := existingMap[item.ID]; !found {
				return utils.ErrNotFound
			}
		}

		result = make([]db.Todo, len(cleanUpdates))
		for i, item := range cleanUpdates {
			todo := existingMap[item.ID]
			if item.Message != nil {
				todo.Message = *item.Message
			}
			if item.Check != nil {
				todo.Check = *item.Check
			}
			if err := txRepo.Save(ctx, &todo); err != nil {
				return utils.ErrInternalServer
			}
			result[i] = todo
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteByID removes a single Todo by ID verifying ownership.
func (s *TodoService) DeleteByID(ctx context.Context, id, userID uint) (int64, error) {
	if id == 0 {
		return 0, utils.ErrBadRequest
	}

	rowsAffected, err := s.repo.DeleteByID(ctx, id, userID)
	if err != nil {
		return 0, utils.ErrInternalServer
	}

	return rowsAffected, nil
}

// DeleteBulk removes multiple Todos by IDs verifying ownership.
func (s *TodoService) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	if len(ids) == 0 {
		return 0, utils.ErrBadRequest
	}

	rowsAffected, err := s.repo.DeleteBulk(ctx, ids, userID)
	if err != nil {
		return 0, utils.ErrInternalServer
	}

	return rowsAffected, nil
}
