package business

import (
	"context"
	"strings"

	"go-gin-todo/src/shared"
)

// TodoService encapsulates business logic and coordinates repository operations.
type TodoService struct {
	repo ITodoRepository
}

// NewTodoService constructs a new TodoService with an injected repository interface.
func NewTodoService(repo ITodoRepository) *TodoService {
	return &TodoService{
		repo: repo,
	}
}

// CreateOne validates and creates a single Todo item for a user.
func (s *TodoService) CreateOne(ctx context.Context, userID uint, message string, check bool) (*Todo, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, shared.ErrBadRequest
	}

	todo := &Todo{
		UserID:  userID,
		Message: message,
		Check:   check,
	}

	if err := s.repo.Create(ctx, todo); err != nil {
		return nil, shared.ErrInternalServer
	}

	return todo, nil
}

// CreateBulk validates and creates multiple Todo items for a user.
func (s *TodoService) CreateBulk(ctx context.Context, userID uint, items []Todo) ([]Todo, error) {
	if len(items) == 0 {
		return nil, shared.ErrBadRequest
	}

	todos := make([]Todo, 0, len(items))
	for _, item := range items {
		msg := strings.TrimSpace(item.Message)
		if msg == "" {
			return nil, shared.ErrBadRequest
		}
		todos = append(todos, Todo{
			UserID:  userID,
			Message: msg,
			Check:   item.Check,
		})
	}

	if err := s.repo.CreateBulk(ctx, &todos); err != nil {
		return nil, shared.ErrInternalServer
	}

	return todos, nil
}

// GetByID retrieves a single Todo by ID verifying ownership.
func (s *TodoService) GetByID(ctx context.Context, id, userID uint) (*Todo, error) {
	if id == 0 {
		return nil, shared.ErrBadRequest
	}

	item, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	return item, nil
}

// GetBulk retrieves paginated and filtered Todos for a user ordered by id DESC.
func (s *TodoService) GetBulk(ctx context.Context, userID uint, search string, checkFilter *bool, page, limit int) ([]Todo, shared.Pagination, error) {
	items, pagination, err := s.repo.GetBulk(ctx, userID, search, checkFilter, page, limit)
	if err != nil {
		return nil, shared.Pagination{}, shared.ErrInternalServer
	}

	return items, pagination, nil
}

// UpdateByID updates message and/or check status of a user-owned Todo.
func (s *TodoService) UpdateByID(ctx context.Context, id, userID uint, message *string, check *bool) (*Todo, error) {
	if id == 0 || (message == nil && check == nil) {
		return nil, shared.ErrBadRequest
	}

	var trimmedMsg *string
	if message != nil {
		msg := strings.TrimSpace(*message)
		if msg == "" {
			return nil, shared.ErrBadRequest
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
		return nil, shared.ErrInternalServer
	}

	return item, nil
}

// UpdateBulk validates and atomically updates multiple Todos owned by the user via the repository.
func (s *TodoService) UpdateBulk(ctx context.Context, userID uint, updates []TodoUpdateItem) ([]Todo, error) {
	if len(updates) == 0 {
		return nil, shared.ErrBadRequest
	}

	cleanUpdates := make([]TodoUpdateItem, len(updates))
	for i, item := range updates {
		if item.ID == 0 {
			return nil, shared.ErrBadRequest
		}
		var cleanMsg *string
		if item.Message != nil {
			msg := strings.TrimSpace(*item.Message)
			if msg == "" {
				return nil, shared.ErrBadRequest
			}
			cleanMsg = &msg
		}
		cleanUpdates[i] = TodoUpdateItem{
			ID:      item.ID,
			Message: cleanMsg,
			Check:   item.Check,
		}
	}

	return s.repo.UpdateBulk(ctx, userID, cleanUpdates)
}

// DeleteByID removes a single Todo by ID verifying ownership.
func (s *TodoService) DeleteByID(ctx context.Context, id, userID uint) (int64, error) {
	if id == 0 {
		return 0, shared.ErrBadRequest
	}

	rowsAffected, err := s.repo.DeleteByID(ctx, id, userID)
	if err != nil {
		return 0, shared.ErrInternalServer
	}

	return rowsAffected, nil
}

// DeleteBulk removes multiple Todos by IDs verifying ownership.
func (s *TodoService) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	if len(ids) == 0 {
		return 0, shared.ErrBadRequest
	}

	rowsAffected, err := s.repo.DeleteBulk(ctx, ids, userID)
	if err != nil {
		return 0, shared.ErrInternalServer
	}

	return rowsAffected, nil
}
