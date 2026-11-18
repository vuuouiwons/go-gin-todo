package app

import (
	"context"
	"strings"
)

// TodolistCommands handles write operations and business rules for Todolists.
type TodolistCommands struct {
	repo TodolistRepository
	tx   Transactor
}

// NewTodolistCommands constructs a new TodolistCommands handler.
func NewTodolistCommands(repo TodolistRepository, tx Transactor) *TodolistCommands {
	return &TodolistCommands{repo: repo, tx: tx}
}

// CreateOne validates and creates a single Todolist for a user inside a transaction.
func (c *TodolistCommands) CreateOne(ctx context.Context, userID uint, title string) (*Todolist, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrBadRequest
	}

	return RunInTx(ctx, c.tx, func(txCtx context.Context) (*Todolist, error) {
		todolist := &Todolist{
			UserID: userID,
			Title:  title,
		}

		if err := c.repo.Create(txCtx, todolist); err != nil {
			return nil, ErrInternalServer
		}

		return todolist, nil
	})
}

// CreateBulk validates and creates multiple Todolists for a user inside a transaction.
func (c *TodolistCommands) CreateBulk(ctx context.Context, userID uint, titles []string) ([]Todolist, error) {
	if len(titles) == 0 {
		return nil, ErrBadRequest
	}

	items := make([]Todolist, 0, len(titles))
	for _, t := range titles {
		cleanTitle := strings.TrimSpace(t)
		if cleanTitle == "" {
			return nil, ErrBadRequest
		}
		items = append(items, Todolist{
			UserID: userID,
			Title:  cleanTitle,
		})
	}

	return RunInTx(ctx, c.tx, func(txCtx context.Context) ([]Todolist, error) {
		created, err := c.repo.CreateBulk(txCtx, items)
		if err != nil {
			return nil, ErrInternalServer
		}
		return created, nil
	})
}

// UpdateByID updates the title of a user-owned Todolist inside a transaction.
func (c *TodolistCommands) UpdateByID(ctx context.Context, id, userID uint, title string) (*Todolist, error) {
	title = strings.TrimSpace(title)
	if id == 0 || title == "" {
		return nil, ErrBadRequest
	}

	return RunInTx(ctx, c.tx, func(txCtx context.Context) (*Todolist, error) {
		todolist, err := c.repo.GetByID(txCtx, id, userID)
		if err != nil {
			return nil, err
		}

		todolist.Title = title
		if err := c.repo.Update(txCtx, todolist); err != nil {
			return nil, ErrInternalServer
		}

		return todolist, nil
	})
}

// UpdateBulk atomically updates multiple Todolists owned by the user via RunInTx.
func (c *TodolistCommands) UpdateBulk(ctx context.Context, userID uint, updates []Todolist) ([]Todolist, error) {
	if len(updates) == 0 {
		return nil, ErrBadRequest
	}

	cleanUpdates := make([]Todolist, len(updates))
	for i, item := range updates {
		title := strings.TrimSpace(item.Title)
		if item.ID == 0 || title == "" {
			return nil, ErrBadRequest
		}
		cleanUpdates[i] = Todolist{
			ID:     item.ID,
			UserID: userID,
			Title:  title,
		}
	}

	return RunInTx(ctx, c.tx, func(txCtx context.Context) ([]Todolist, error) {
		return c.repo.UpdateBulk(txCtx, userID, cleanUpdates)
	})
}

// DeleteByID deletes a single Todolist by ID if owned by the user inside a transaction.
func (c *TodolistCommands) DeleteByID(ctx context.Context, id, userID uint) (int64, error) {
	if id == 0 {
		return 0, ErrBadRequest
	}
	return RunInTx(ctx, c.tx, func(txCtx context.Context) (int64, error) {
		return c.repo.DeleteByID(txCtx, id, userID)
	})
}

// DeleteBulk deletes multiple Todolists by IDs if owned by the user inside a transaction.
func (c *TodolistCommands) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	if len(ids) == 0 {
		return 0, ErrBadRequest
	}
	return RunInTx(ctx, c.tx, func(txCtx context.Context) (int64, error) {
		return c.repo.DeleteBulk(txCtx, ids, userID)
	})
}
