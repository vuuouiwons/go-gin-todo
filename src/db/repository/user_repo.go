package repository

import (
	"context"
	"errors"

	"go-gin-todo/src/db"
	"go-gin-todo/src/utils"
	"gorm.io/gorm"
)

// UserRepository handles database operations and reusable queries for User records.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository constructs a new UserRepository.
func NewUserRepository(database *gorm.DB) *UserRepository {
	return &UserRepository{db: database}
}

// WithTx returns a new UserRepository instance bound to an active transaction.
func (r *UserRepository) WithTx(tx *gorm.DB) *UserRepository {
	if tx == nil {
		return r
	}
	return &UserRepository{db: tx}
}

// Create inserts a new user record.
func (r *UserRepository) Create(ctx context.Context, user *db.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByEmail retrieves a user by email, returning utils.ErrNotFound if absent.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*db.User, error) {
	var user db.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByID retrieves a user by ID, returning utils.ErrNotFound if absent.
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*db.User, error) {
	var user db.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
