package business

import (
	"context"
	"strings"

	"go-gin-todo/src/shared"
)

// TodolistService encapsulates business logic and coordinates repository operations.
type TodolistService struct {
	repo ITodolistRepository
}

// NewTodolistService constructs a new TodolistService with an injected repository interface.
func NewTodolistService(repo ITodolistRepository) *TodolistService {
	return &TodolistService{
		repo: repo,
	}
}

// CreateOne validates and persists a single Todolist for a user.
func (s *TodolistService) CreateOne(ctx context.Context, userID uint, title string) (*Todolist, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, shared.ErrBadRequest
	}

	todolist := &Todolist{
		UserID: userID,
		Title:  title,
	}

	if err := s.repo.Create(ctx, todolist); err != nil {
		return nil, shared.ErrInternalServer
	}

	return todolist, nil
}

// CreateBulk validates and creates multiple Todolists for a user.
func (s *TodolistService) CreateBulk(ctx context.Context, userID uint, titles []string) ([]Todolist, error) {
	if len(titles) == 0 {
		return nil, shared.ErrBadRequest
	}

	items := make([]Todolist, 0, len(titles))
	for _, t := range titles {
		cleanTitle := strings.TrimSpace(t)
		if cleanTitle == "" {
			return nil, shared.ErrBadRequest
		}
		items = append(items, Todolist{
			UserID: userID,
			Title:  cleanTitle,
		})
	}

	if err := s.repo.CreateBulk(ctx, &items); err != nil {
		return nil, shared.ErrInternalServer
	}

	return items, nil
}

// GetByID retrieves a single Todolist by ID verifying user ownership.
func (s *TodolistService) GetByID(ctx context.Context, id, userID uint) (*Todolist, error) {
	if id == 0 {
		return nil, shared.ErrBadRequest
	}

	item, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	return item, nil
}

// GetBulk retrieves paginated and filtered Todolists for a user ordered by id DESC.
func (s *TodolistService) GetBulk(ctx context.Context, userID uint, search string, page, limit int) ([]Todolist, shared.Pagination, error) {
	items, pagination, err := s.repo.GetBulk(ctx, userID, search, page, limit)
	if err != nil {
		return nil, shared.Pagination{}, shared.ErrInternalServer
	}

	return items, pagination, nil
}

// UpdateByID updates the title of a single user-owned Todolist.
func (s *TodolistService) UpdateByID(ctx context.Context, id, userID uint, title string) (*Todolist, error) {
	title = strings.TrimSpace(title)
	if id == 0 || title == "" {
		return nil, shared.ErrBadRequest
	}

	item, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	item.Title = title
	if err := s.repo.Save(ctx, item); err != nil {
		return nil, shared.ErrInternalServer
	}

	return item, nil
}

// UpdateBulk validates and atomically updates multiple Todolists owned by the user via the repository.
func (s *TodolistService) UpdateBulk(ctx context.Context, userID uint, updates []Todolist) ([]Todolist, error) {
	if len(updates) == 0 {
		return nil, shared.ErrBadRequest
	}

	cleanUpdates := make([]Todolist, len(updates))
	for i, u := range updates {
		title := strings.TrimSpace(u.Title)
		if u.ID == 0 || title == "" {
			return nil, shared.ErrBadRequest
		}
		cleanUpdates[i] = Todolist{
			ID:     u.ID,
			UserID: userID,
			Title:  title,
		}
	}

	return s.repo.UpdateBulk(ctx, userID, cleanUpdates)
}

// DeleteByID removes a single Todolist by ID verifying ownership.
func (s *TodolistService) DeleteByID(ctx context.Context, id, userID uint) (int64, error) {
	if id == 0 {
		return 0, shared.ErrBadRequest
	}

	rowsAffected, err := s.repo.DeleteByID(ctx, id, userID)
	if err != nil {
		return 0, shared.ErrInternalServer
	}

	return rowsAffected, nil
}

// DeleteBulk removes multiple Todolists by IDs verifying ownership.
func (s *TodolistService) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	if len(ids) == 0 {
		return 0, shared.ErrBadRequest
	}

	rowsAffected, err := s.repo.DeleteBulk(ctx, ids, userID)
	if err != nil {
		return 0, shared.ErrInternalServer
	}

	return rowsAffected, nil
}
