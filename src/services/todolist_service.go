package services

import (
	"context"
	"strings"

	"go-gin-todo/src/db"
	"go-gin-todo/src/db/repository"
	"go-gin-todo/src/utils"
	"gorm.io/gorm"
)

// TodolistService encapsulates business logic and coordinates repository operations.
type TodolistService struct {
	db   *gorm.DB
	repo *repository.TodolistRepository
}

// NewTodolistService constructs a new TodolistService with a reusable TodolistRepository and DB handle for transactions.
func NewTodolistService(database *gorm.DB, repo *repository.TodolistRepository) *TodolistService {
	return &TodolistService{
		db:   database,
		repo: repo,
	}
}

// CreateOne validates and persists a single Todolist for a user.
func (s *TodolistService) CreateOne(ctx context.Context, userID uint, title string) (*db.Todolist, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, utils.ErrBadRequest
	}

	todolist := &db.Todolist{
		UserID: userID,
		Title:  title,
	}

	if err := s.repo.Create(ctx, todolist); err != nil {
		return nil, utils.ErrInternalServer
	}

	return todolist, nil
}

// CreateBulk validates and creates multiple Todolists for a user.
func (s *TodolistService) CreateBulk(ctx context.Context, userID uint, titles []string) ([]db.Todolist, error) {
	if len(titles) == 0 {
		return nil, utils.ErrBadRequest
	}

	items := make([]db.Todolist, 0, len(titles))
	for _, t := range titles {
		cleanTitle := strings.TrimSpace(t)
		if cleanTitle == "" {
			return nil, utils.ErrBadRequest
		}
		items = append(items, db.Todolist{
			UserID: userID,
			Title:  cleanTitle,
		})
	}

	if err := s.repo.CreateBulk(ctx, &items); err != nil {
		return nil, utils.ErrInternalServer
	}

	return items, nil
}

// GetByID retrieves a single Todolist by ID verifying user ownership.
func (s *TodolistService) GetByID(ctx context.Context, id, userID uint) (*db.Todolist, error) {
	if id == 0 {
		return nil, utils.ErrBadRequest
	}

	item, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	return item, nil
}

// GetBulk retrieves paginated and filtered Todolists for a user ordered by id DESC.
func (s *TodolistService) GetBulk(ctx context.Context, userID uint, search string, page, limit int) ([]db.Todolist, db.Pagination, error) {
	items, pagination, err := s.repo.GetBulk(ctx, userID, search, page, limit)
	if err != nil {
		return nil, db.Pagination{}, utils.ErrInternalServer
	}

	return items, pagination, nil
}

// UpdateByID updates the title of a single user-owned Todolist.
func (s *TodolistService) UpdateByID(ctx context.Context, id, userID uint, title string) (*db.Todolist, error) {
	title = strings.TrimSpace(title)
	if id == 0 || title == "" {
		return nil, utils.ErrBadRequest
	}

	item, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	item.Title = title
	if err := s.repo.Save(ctx, item); err != nil {
		return nil, utils.ErrInternalServer
	}

	return item, nil
}

// UpdateBulk atomically updates multiple Todolists owned by the user using a GORM database transaction.
func (s *TodolistService) UpdateBulk(ctx context.Context, userID uint, updates []db.Todolist) ([]db.Todolist, error) {
	if len(updates) == 0 {
		return nil, utils.ErrBadRequest
	}

	cleanUpdates := make([]db.Todolist, len(updates))
	ids := make([]uint, len(updates))
	for i, u := range updates {
		title := strings.TrimSpace(u.Title)
		if u.ID == 0 || title == "" {
			return nil, utils.ErrBadRequest
		}
		cleanUpdates[i] = db.Todolist{
			ID:     u.ID,
			UserID: userID,
			Title:  title,
		}
		ids[i] = u.ID
	}

	var result []db.Todolist

	// Typical GORM transaction pattern with repository scoping via WithTx
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		existing, err := txRepo.FindByIDs(ctx, userID, ids)
		if err != nil {
			return utils.ErrInternalServer
		}

		existingMap := make(map[uint]db.Todolist, len(existing))
		for _, item := range existing {
			existingMap[item.ID] = item
		}

		for _, u := range cleanUpdates {
			if _, found := existingMap[u.ID]; !found {
				return utils.ErrNotFound
			}
		}

		result = make([]db.Todolist, len(cleanUpdates))
		for i, u := range cleanUpdates {
			item := existingMap[u.ID]
			item.Title = u.Title
			if err := txRepo.Save(ctx, &item); err != nil {
				return utils.ErrInternalServer
			}
			result[i] = item
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteByID removes a single Todolist by ID verifying ownership.
func (s *TodolistService) DeleteByID(ctx context.Context, id, userID uint) (int64, error) {
	if id == 0 {
		return 0, utils.ErrBadRequest
	}

	rowsAffected, err := s.repo.DeleteByID(ctx, id, userID)
	if err != nil {
		return 0, utils.ErrInternalServer
	}

	return rowsAffected, nil
}

// DeleteBulk removes multiple Todolists by IDs verifying ownership.
func (s *TodolistService) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	if len(ids) == 0 {
		return 0, utils.ErrBadRequest
	}

	rowsAffected, err := s.repo.DeleteBulk(ctx, ids, userID)
	if err != nil {
		return 0, utils.ErrInternalServer
	}

	return rowsAffected, nil
}
