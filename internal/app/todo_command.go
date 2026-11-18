package app

import (
	"context"
	"strings"
)

// TodoCommands handles write operations and business rules for Todos.
type TodoCommands struct {
	repo TodoRepository
	tx   Transactor
}

// NewTodoCommands constructs a new TodoCommands handler.
func NewTodoCommands(repo TodoRepository, tx Transactor) *TodoCommands {
	return &TodoCommands{repo: repo, tx: tx}
}

// CreateOne validates and creates a single Todo for a user inside a transaction.
func (c *TodoCommands) CreateOne(ctx context.Context, userID uint, message string, check bool) (*Todo, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, ErrBadRequest
	}

	return RunInTx(ctx, c.tx, func(txCtx context.Context) (*Todo, error) {
		todo := &Todo{
			UserID:  userID,
			Message: message,
			Check:   check,
		}

		if err := c.repo.Create(txCtx, todo); err != nil {
			return nil, ErrInternalServer
		}

		return todo, nil
	})
}

// CreateBulk validates and creates multiple Todos for a user inside a transaction.
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

	return RunInTx(ctx, c.tx, func(txCtx context.Context) ([]Todo, error) {
		created, err := c.repo.CreateBulk(txCtx, todos)
		if err != nil {
			return nil, ErrInternalServer
		}
		return created, nil
	})
}

// UpdateByID updates fields of a user-owned Todo inside a transaction.
func (c *TodoCommands) UpdateByID(ctx context.Context, id, userID uint, message *string, check *bool) (*Todo, error) {
	if id == 0 || (message == nil && check == nil) {
		return nil, ErrBadRequest
	}

	var trimmedMsg *string
	if message != nil {
		msg := strings.TrimSpace(*message)
		if msg == "" {
			return nil, ErrBadRequest
		}
		trimmedMsg = &msg
	}

	return RunInTx(ctx, c.tx, func(txCtx context.Context) (*Todo, error) {
		todo, err := c.repo.GetByID(txCtx, id, userID)
		if err != nil {
			return nil, err
		}

		if trimmedMsg != nil {
			todo.Message = *trimmedMsg
		}

		if check != nil {
			todo.Check = *check
		}

		if err := c.repo.Update(txCtx, todo); err != nil {
			return nil, ErrInternalServer
		}

		return todo, nil
	})
}

// UpdateBulk atomically updates multiple Todos owned by the user via RunInTx.
func (c *TodoCommands) UpdateBulk(ctx context.Context, userID uint, updates []TodoUpdateItem) ([]Todo, error) {
	if len(updates) == 0 {
		return nil, ErrBadRequest
	}

	cleanUpdates := make([]TodoUpdateItem, len(updates))
	for i, item := range updates {
		if item.ID == 0 {
			return nil, ErrBadRequest
		}
		var cleanMsg *string
		if item.Message != nil {
			msg := strings.TrimSpace(*item.Message)
			if msg == "" {
				return nil, ErrBadRequest
			}
			cleanMsg = &msg
		}
		cleanUpdates[i] = TodoUpdateItem{
			ID:      item.ID,
			Message: cleanMsg,
			Check:   item.Check,
		}
	}

	return RunInTx(ctx, c.tx, func(txCtx context.Context) ([]Todo, error) {
		return c.repo.UpdateBulk(txCtx, userID, cleanUpdates)
	})
}

// DeleteByID deletes a single Todo by ID if owned by the user inside a transaction.
func (c *TodoCommands) DeleteByID(ctx context.Context, id, userID uint) (int64, error) {
	if id == 0 {
		return 0, ErrBadRequest
	}
	return RunInTx(ctx, c.tx, func(txCtx context.Context) (int64, error) {
		return c.repo.DeleteByID(txCtx, id, userID)
	})
}

// DeleteBulk deletes multiple Todos by IDs if owned by the user inside a transaction.
func (c *TodoCommands) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	if len(ids) == 0 {
		return 0, ErrBadRequest
	}
	return RunInTx(ctx, c.tx, func(txCtx context.Context) (int64, error) {
		return c.repo.DeleteBulk(txCtx, ids, userID)
	})
}
