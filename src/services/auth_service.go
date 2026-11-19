package services

import (
	"context"
	"errors"
	"strings"

	"go-gin-todo/src/db"
	"go-gin-todo/src/utils"
	"gorm.io/gorm"
)

// AuthService handles authentication, user registration, and credential verification.
type AuthService struct {
	db                *gorm.DB
	jwtSecret         string
	jwtExpiresInHours int
}

// NewAuthService constructs a new AuthService instance.
func NewAuthService(database *gorm.DB, jwtSecret string, jwtExpiresInHours int) *AuthService {
	return &AuthService{
		db:                database,
		jwtSecret:         jwtSecret,
		jwtExpiresInHours: jwtExpiresInHours,
	}
}

// Register validates user credentials, ensures email uniqueness, hashes the password, and creates the account.
func (s *AuthService) Register(ctx context.Context, email, password string) (*db.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || len(password) < 6 {
		return nil, utils.ErrBadRequest
	}

	var existing db.User
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&existing).Error
	if err == nil {
		return nil, utils.ErrConflict
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrInternalServer
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, utils.ErrInternalServer
	}

	user := &db.User{
		Email:    email,
		Password: hashedPassword,
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, utils.ErrInternalServer
	}

	return user, nil
}

// Login validates user email and password, returning the user and a signed JWT token.
func (s *AuthService) Login(ctx context.Context, email, password string) (*db.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return nil, "", utils.ErrUnauthorized
	}

	var user db.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", utils.ErrUnauthorized
		}
		return nil, "", utils.ErrInternalServer
	}

	if !utils.CheckPassword(user.Password, password) {
		return nil, "", utils.ErrUnauthorized
	}

	token, err := utils.GenerateToken(user.ID, user.Email, s.jwtSecret, s.jwtExpiresInHours)
	if err != nil {
		return nil, "", utils.ErrInternalServer
	}

	return &user, token, nil
}
