package usecase

import (
	"context"
	"strings"

	"go-gin-todo/internal/domain"
)

type TodoUsecase interface {
	GetBulk(ctx context.Context, userID uint, search string, checkFilter *bool, page, limit int) ([]domain.Todo, domain.Pagination, error)
	GetByID(ctx context.Context, id uint, userID uint) (*domain.Todo, error)
	CreateOne(ctx context.Context, userID uint, message string, check bool) (*domain.Todo, error)
	CreateBulk(ctx context.Context, userID uint, items []domain.Todo) ([]domain.Todo, error)
	UpdateByID(ctx context.Context, id uint, userID uint, message *string, check *bool) (*domain.Todo, error)
	UpdateBulk(ctx context.Context, userID uint, updates []domain.TodoUpdateItem) ([]domain.Todo, error)
	DeleteByID(ctx context.Context, id uint, userID uint) (int64, error)
	DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error)
}

type todoUsecase struct {
	repo domain.TodoRepository
}

func NewTodoUsecase(repo domain.TodoRepository) TodoUsecase {
	return &todoUsecase{repo: repo}
}

func (u *todoUsecase) GetBulk(ctx context.Context, userID uint, search string, checkFilter *bool, page, limit int) ([]domain.Todo, domain.Pagination, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	items, total, err := u.repo.GetBulk(ctx, userID, search, checkFilter, page, limit)
	if err != nil {
		return nil, domain.Pagination{}, err
	}

	pagination := domain.NewPagination(page, limit, total)
	return items, pagination, nil
}

func (u *todoUsecase) GetByID(ctx context.Context, id uint, userID uint) (*domain.Todo, error) {
	if id == 0 {
		return nil, domain.ErrBadRequest
	}
	return u.repo.GetByID(ctx, id, userID)
}

func (u *todoUsecase) CreateOne(ctx context.Context, userID uint, message string, check bool) (*domain.Todo, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, domain.ErrBadRequest
	}

	todo := &domain.Todo{
		UserID:  userID,
		Message: message,
		Check:   check,
	}

	if err := u.repo.Create(ctx, todo); err != nil {
		return nil, domain.ErrInternalServer
	}

	return todo, nil
}

func (u *todoUsecase) CreateBulk(ctx context.Context, userID uint, items []domain.Todo) ([]domain.Todo, error) {
	if len(items) == 0 {
		return nil, domain.ErrBadRequest
	}

	var todos []domain.Todo
	for _, item := range items {
		msg := strings.TrimSpace(item.Message)
		if msg == "" {
			return nil, domain.ErrBadRequest
		}
		todos = append(todos, domain.Todo{
			UserID:  userID,
			Message: msg,
			Check:   item.Check,
		})
	}

	created, err := u.repo.CreateBulk(ctx, todos)
	if err != nil {
		return nil, domain.ErrInternalServer
	}

	return created, nil
}

func (u *todoUsecase) UpdateByID(ctx context.Context, id uint, userID uint, message *string, check *bool) (*domain.Todo, error) {
	if id == 0 || (message == nil && check == nil) {
		return nil, domain.ErrBadRequest
	}

	todo, err := u.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if message != nil {
		msg := strings.TrimSpace(*message)
		if msg == "" {
			return nil, domain.ErrBadRequest
		}
		todo.Message = msg
	}

	if check != nil {
		todo.Check = *check
	}

	if err := u.repo.Update(ctx, todo); err != nil {
		return nil, domain.ErrInternalServer
	}

	return todo, nil
}

func (u *todoUsecase) UpdateBulk(ctx context.Context, userID uint, updates []domain.TodoUpdateItem) ([]domain.Todo, error) {
	if len(updates) == 0 {
		return nil, domain.ErrBadRequest
	}

	for _, item := range updates {
		if item.ID == 0 {
			return nil, domain.ErrBadRequest
		}
		if item.Message != nil && strings.TrimSpace(*item.Message) == "" {
			return nil, domain.ErrBadRequest
		}
	}

	return u.repo.UpdateBulk(ctx, userID, updates)
}

func (u *todoUsecase) DeleteByID(ctx context.Context, id uint, userID uint) (int64, error) {
	if id == 0 {
		return 0, domain.ErrBadRequest
	}
	return u.repo.DeleteByID(ctx, id, userID)
}

func (u *todoUsecase) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	if len(ids) == 0 {
		return 0, domain.ErrBadRequest
	}
	return u.repo.DeleteBulk(ctx, ids, userID)
}
