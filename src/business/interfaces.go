package business

import (
	"context"

	"go-gin-todo/src/shared"
)

// IUserRepository defines persistence operations required for User domain models.
type IUserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uint) (*User, error)
}

// ITodolistRepository defines persistence operations required for Todolist domain models.
type ITodolistRepository interface {
	Create(ctx context.Context, item *Todolist) error
	CreateBulk(ctx context.Context, items *[]Todolist) error
	GetByID(ctx context.Context, id, userID uint) (*Todolist, error)
	GetBulk(ctx context.Context, userID uint, search string, page, limit int) ([]Todolist, shared.Pagination, error)
	FindByIDs(ctx context.Context, userID uint, ids []uint) ([]Todolist, error)
	Save(ctx context.Context, item *Todolist) error
	UpdateBulk(ctx context.Context, userID uint, updates []Todolist) ([]Todolist, error)
	DeleteByID(ctx context.Context, id, userID uint) (int64, error)
	DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error)
}

// ITodoRepository defines persistence operations required for Todo domain models.
type ITodoRepository interface {
	Create(ctx context.Context, todo *Todo) error
	CreateBulk(ctx context.Context, todos *[]Todo) error
	GetByID(ctx context.Context, id, userID uint) (*Todo, error)
	GetBulk(ctx context.Context, userID uint, search string, checkFilter *bool, page, limit int) ([]Todo, shared.Pagination, error)
	FindByIDs(ctx context.Context, userID uint, ids []uint) ([]Todo, error)
	Save(ctx context.Context, todo *Todo) error
	UpdateBulk(ctx context.Context, userID uint, updates []TodoUpdateItem) ([]Todo, error)
	DeleteByID(ctx context.Context, id, userID uint) (int64, error)
	DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error)
}
