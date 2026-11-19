package data

import (
	"context"
	"errors"

	"go-gin-todo/src/business"
	"go-gin-todo/src/shared"
	"gorm.io/gorm"
)

// UserRepository implements business.IUserRepository using GORM.
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
func (r *UserRepository) Create(ctx context.Context, user *business.User) error {
	entity := UserToEntity(user)
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return err
	}
	user.ID = entity.ID
	user.CreatedAt = entity.CreatedAt
	user.UpdatedAt = entity.UpdatedAt
	return nil
}

// GetByEmail retrieves a user by email, returning shared.ErrNotFound if absent.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*business.User, error) {
	var entity UserEntity
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return UserToDomain(&entity), nil
}

// GetByID retrieves a user by ID, returning shared.ErrNotFound if absent.
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*business.User, error) {
	var entity UserEntity
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return UserToDomain(&entity), nil
}
