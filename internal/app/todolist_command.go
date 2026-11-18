package app

import (
	"context"
	"strings"
)

// TodolistCommands handles write operations and business rules for Todolists.
type TodolistCommands struct {
	repo TodolistRepository
}

// NewTodolistCommands constructs a new TodolistCommands handler.
func NewTodolistCommands(repo TodolistRepository) *TodolistCommands {
	return &TodolistCommands{repo: repo}
}

// CreateOne validates and creates a single Todolist for a user.
func (c *TodolistCommands) CreateOne(ctx context.Context, userID uint, title string) (*Todolist, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrBadRequest
	}

	todolist := &Todolist{
		UserID: userID,
		Title:  title,
	}

	if err := c.repo.Create(ctx, todolist); err != nil {
		return nil, ErrInternalServer
	}

	return todolist, nil
}

// CreateBulk validates and creates multiple Todolists for a user.
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

	created, err := c.repo.CreateBulk(ctx, items)
	if err != nil {
		return nil, ErrInternalServer
	}

	return created, nil
}

// UpdateByID updates the title of a user-owned Todolist.
func (c *TodolistCommands) UpdateByID(ctx context.Context, id, userID uint, title string) (*Todolist, error) {
	title = strings.TrimSpace(title)
	if id == 0 || title == "" {
		return nil, ErrBadRequest
	}

	todolist, err := c.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	todolist.Title = title
	if err := c.repo.Update(ctx, todolist); err != nil {
		return nil, ErrInternalServer
	}

	return todolist, nil
}

// UpdateBulk updates multiple Todolists owned by the user (participates in the active transaction in ctx).
func (c *TodolistCommands) UpdateBulk(ctx context.Context, userID uint, updates []Todolist) ([]Todolist, error) {
	if len(updates) == 0 {
		return nil, ErrBadRequest
	}

	for _, item := range updates {
		if item.ID == 0 || strings.TrimSpace(item.Title) == "" {
			return nil, ErrBadRequest
		}
	}

	updated := make([]Todolist, 0, len(updates))
	for _, item := range updates {
		existing, err := c.repo.GetByID(ctx, item.ID, userID)
		if err != nil {
			return nil, err
		}
		existing.Title = strings.TrimSpace(item.Title)
		if err := c.repo.Update(ctx, existing); err != nil {
			return nil, err
		}
		updated = append(updated, *existing)
	}

	return updated, nil
}

// DeleteByID deletes a single Todolist by ID if owned by the user.
func (c *TodolistCommands) DeleteByID(ctx context.Context, id, userID uint) (int64, error) {
	if id == 0 {
		return 0, ErrBadRequest
	}
	return c.repo.DeleteByID(ctx, id, userID)
}

// DeleteBulk deletes multiple Todolists by IDs if owned by the user.
func (c *TodolistCommands) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	if len(ids) == 0 {
		return 0, ErrBadRequest
	}
	return c.repo.DeleteBulk(ctx, ids, userID)
}
