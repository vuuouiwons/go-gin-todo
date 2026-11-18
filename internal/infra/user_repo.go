package infra

import (
	"context"
	"errors"

	"go-gin-todo/internal/app"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository constructs a GORM implementation of app.UserRepository.
func NewUserRepository(db *gorm.DB) app.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *app.User) error {
	return dbFromContext(ctx, r.db).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*app.User, error) {
	var user app.User
	if err := dbFromContext(ctx, r.db).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, app.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*app.User, error) {
	var user app.User
	if err := dbFromContext(ctx, r.db).Where(&app.User{Email: email}).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, app.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
