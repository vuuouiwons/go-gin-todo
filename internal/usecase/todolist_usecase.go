package usecase

import (
	"context"
	"strings"

	"go-gin-todo/internal/domain"
)

type TodolistUsecase interface {
	GetBulk(ctx context.Context, userID uint, search string, page, pageSize int) ([]domain.Todolist, domain.Pagination, error)
	GetByID(ctx context.Context, id uint, userID uint) (*domain.Todolist, error)
	CreateOne(ctx context.Context, userID uint, title string) (*domain.Todolist, error)
	CreateBulk(ctx context.Context, userID uint, titles []string) ([]domain.Todolist, error)
	UpdateByID(ctx context.Context, id uint, userID uint, title string) (*domain.Todolist, error)
	UpdateBulk(ctx context.Context, userID uint, updates []domain.Todolist) ([]domain.Todolist, error)
	DeleteByID(ctx context.Context, id uint, userID uint) (int64, error)
	DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error)
}

type todolistUsecase struct {
	repo domain.TodolistRepository
}

func NewTodolistUsecase(repo domain.TodolistRepository) TodolistUsecase {
	return &todolistUsecase{repo: repo}
}

func (u *todolistUsecase) GetBulk(ctx context.Context, userID uint, search string, page, pageSize int) ([]domain.Todolist, domain.Pagination, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	items, total, err := u.repo.GetBulk(ctx, userID, search, page, pageSize)
	if err != nil {
		return nil, domain.Pagination{}, err
	}

	pagination := domain.NewPagination(page, pageSize, total)
	return items, pagination, nil
}

func (u *todolistUsecase) GetByID(ctx context.Context, id uint, userID uint) (*domain.Todolist, error) {
	if id == 0 {
		return nil, domain.ErrBadRequest
	}
	return u.repo.GetByID(ctx, id, userID)
}

func (u *todolistUsecase) CreateOne(ctx context.Context, userID uint, title string) (*domain.Todolist, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, domain.ErrBadRequest
	}

	todolist := &domain.Todolist{
		UserID: userID,
		Title:  title,
	}

	if err := u.repo.Create(ctx, todolist); err != nil {
		return nil, domain.ErrInternalServer
	}

	return todolist, nil
}

func (u *todolistUsecase) CreateBulk(ctx context.Context, userID uint, titles []string) ([]domain.Todolist, error) {
	if len(titles) == 0 {
		return nil, domain.ErrBadRequest
	}

	var items []domain.Todolist
	for _, t := range titles {
		cleanTitle := strings.TrimSpace(t)
		if cleanTitle == "" {
			return nil, domain.ErrBadRequest
		}
		items = append(items, domain.Todolist{
			UserID: userID,
			Title:  cleanTitle,
		})
	}

	created, err := u.repo.CreateBulk(ctx, items)
	if err != nil {
		return nil, domain.ErrInternalServer
	}

	return created, nil
}

func (u *todolistUsecase) UpdateByID(ctx context.Context, id uint, userID uint, title string) (*domain.Todolist, error) {
	title = strings.TrimSpace(title)
	if id == 0 || title == "" {
		return nil, domain.ErrBadRequest
	}

	todolist, err := u.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	todolist.Title = title
	if err := u.repo.Update(ctx, todolist); err != nil {
		return nil, domain.ErrInternalServer
	}

	return todolist, nil
}

func (u *todolistUsecase) UpdateBulk(ctx context.Context, userID uint, updates []domain.Todolist) ([]domain.Todolist, error) {
	if len(updates) == 0 {
		return nil, domain.ErrBadRequest
	}

	for _, item := range updates {
		if item.ID == 0 || strings.TrimSpace(item.Title) == "" {
			return nil, domain.ErrBadRequest
		}
	}

	return u.repo.UpdateBulk(ctx, userID, updates)
}

func (u *todolistUsecase) DeleteByID(ctx context.Context, id uint, userID uint) (int64, error) {
	if id == 0 {
		return 0, domain.ErrBadRequest
	}
	return u.repo.DeleteByID(ctx, id, userID)
}

func (u *todolistUsecase) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	if len(ids) == 0 {
		return 0, domain.ErrBadRequest
	}
	return u.repo.DeleteBulk(ctx, ids, userID)
}
