package services

import (
	"context"
	"errors"
	"strings"

	"go-gin-todo/src/db"
	"go-gin-todo/src/db/repository"
	"go-gin-todo/src/utils"
)

// AuthService handles authentication, user registration, and credential verification.
type AuthService struct {
	userRepo          *repository.UserRepository
	jwtSecret         string
	jwtExpiresInHours int
}

// NewAuthService constructs a new AuthService instance.
func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, jwtExpiresInHours int) *AuthService {
	return &AuthService{
		userRepo:          userRepo,
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

	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, utils.ErrConflict
	}
	if err != nil && !errors.Is(err, utils.ErrNotFound) {
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

	if err := s.userRepo.Create(ctx, user); err != nil {
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

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
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

	return user, token, nil
}
