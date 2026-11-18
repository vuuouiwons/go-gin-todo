package app

import (
	"context"
	"strings"
)

// TodoCommands handles write operations and business rules for Todos.
type TodoCommands struct {
	repo TodoRepository
}

// NewTodoCommands constructs a new TodoCommands handler.
func NewTodoCommands(repo TodoRepository) *TodoCommands {
	return &TodoCommands{repo: repo}
}

// CreateOne validates and creates a single Todo for a user.
func (c *TodoCommands) CreateOne(ctx context.Context, userID uint, message string, check bool) (*Todo, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, ErrBadRequest
	}

	todo := &Todo{
		UserID:  userID,
		Message: message,
		Check:   check,
	}

	if err := c.repo.Create(ctx, todo); err != nil {
		return nil, ErrInternalServer
	}

	return todo, nil
}

// CreateBulk validates and creates multiple Todos for a user.
func (c *TodoCommands) CreateBulk(ctx context.Context, userID uint, items []Todo) ([]Todo, error) {
	if len(items) == 0 {
		return nil, ErrBadRequest
	}

	todos := make([]Todo, 0, len(items))
	for _, item := range items {
		msg := strings.TrimSpace(item.Message)
		if msg == "" {
			return nil, ErrBadRequest
		}
		todos = append(todos, Todo{
			UserID:  userID,
			Message: msg,
			Check:   item.Check,
		})
	}

	created, err := c.repo.CreateBulk(ctx, todos)
	if err != nil {
		return nil, ErrInternalServer
	}

	return created, nil
}

// UpdateByID updates fields of a user-owned Todo.
func (c *TodoCommands) UpdateByID(ctx context.Context, id, userID uint, message *string, check *bool) (*Todo, error) {
	if id == 0 || (message == nil && check == nil) {
		return nil, ErrBadRequest
	}

	todo, err := c.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if message != nil {
		msg := strings.TrimSpace(*message)
		if msg == "" {
			return nil, ErrBadRequest
		}
		todo.Message = msg
	}

	if check != nil {
		todo.Check = *check
	}

	if err := c.repo.Update(ctx, todo); err != nil {
		return nil, ErrInternalServer
	}

	return todo, nil
}

// UpdateBulk updates multiple Todos owned by the user (participates in the active transaction in ctx).
func (c *TodoCommands) UpdateBulk(ctx context.Context, userID uint, updates []TodoUpdateItem) ([]Todo, error) {
	if len(updates) == 0 {
		return nil, ErrBadRequest
	}

	for _, item := range updates {
		if item.ID == 0 {
			return nil, ErrBadRequest
		}
		if item.Message != nil && strings.TrimSpace(*item.Message) == "" {
			return nil, ErrBadRequest
		}
	}

	updated := make([]Todo, 0, len(updates))
	for _, item := range updates {
		existing, err := c.repo.GetByID(ctx, item.ID, userID)
		if err != nil {
			return nil, err
		}
		if item.Message != nil {
			existing.Message = strings.TrimSpace(*item.Message)
		}
		if item.Check != nil {
			existing.Check = *item.Check
		}
		if err := c.repo.Update(ctx, existing); err != nil {
			return nil, err
		}
		updated = append(updated, *existing)
	}

	return updated, nil
}

// DeleteByID deletes a single Todo by ID if owned by the user.
func (c *TodoCommands) DeleteByID(ctx context.Context, id, userID uint) (int64, error) {
	if id == 0 {
		return 0, ErrBadRequest
	}
	return c.repo.DeleteByID(ctx, id, userID)
}

// DeleteBulk deletes multiple Todos by IDs if owned by the user.
func (c *TodoCommands) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	if len(ids) == 0 {
		return 0, ErrBadRequest
	}
	return c.repo.DeleteBulk(ctx, ids, userID)
}
