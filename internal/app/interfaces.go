package app

import (
	"context"

	"gorm.io/gorm"
)

// Transactor executes operations inside an atomic database transaction.
type Transactor interface {
	ExecInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// RunInTx wraps a command operation in an atomic transaction via Transactor.
func RunInTx[T any](ctx context.Context, tx Transactor, fn func(txCtx context.Context) (T, error)) (T, error) {
	var result T
	err := tx.ExecInTransaction(ctx, func(txCtx context.Context) error {
		var fnErr error
		result, fnErr = fn(txCtx)
		return fnErr
	})
	return result, err
}

// Queryer provides direct database query access for Application Queries.
type Queryer interface {
	DB(ctx context.Context) *gorm.DB
}

// UserRepository defines persistence operations for User accounts.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
}

// TodolistRepository defines persistence operations for Todolist entities.
type TodolistRepository interface {
	Create(ctx context.Context, todolist *Todolist) error
	CreateBulk(ctx context.Context, todolists []Todolist) ([]Todolist, error)
	GetByID(ctx context.Context, id, userID uint) (*Todolist, error)
	Update(ctx context.Context, todolist *Todolist) error
	UpdateBulk(ctx context.Context, userID uint, updates []Todolist) ([]Todolist, error)
	DeleteByID(ctx context.Context, id, userID uint) (int64, error)
	DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error)
}

// TodoRepository defines persistence operations for Todo entities.
type TodoRepository interface {
	Create(ctx context.Context, todo *Todo) error
	CreateBulk(ctx context.Context, todos []Todo) ([]Todo, error)
	GetByID(ctx context.Context, id, userID uint) (*Todo, error)
	Update(ctx context.Context, todo *Todo) error
	UpdateBulk(ctx context.Context, userID uint, updates []TodoUpdateItem) ([]Todo, error)
	DeleteByID(ctx context.Context, id, userID uint) (int64, error)
	DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error)
}

// PasswordHasher hashes and verifies user passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) bool
}

// TokenManager generates and validates authentication tokens.
type TokenManager interface {
	Generate(userID uint, email string) (string, error)
	Validate(tokenStr string) (*TokenClaims, error)
}
