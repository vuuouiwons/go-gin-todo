package business

import (
	"context"
	"errors"
	"strings"

	"go-gin-todo/src/shared"
)

// AuthService handles authentication, user registration, and credential verification.
type AuthService struct {
	userRepo          IUserRepository
	jwtSecret         string
	jwtExpiresInHours int
}

// NewAuthService constructs a new AuthService instance with an inverted repository interface.
func NewAuthService(userRepo IUserRepository, jwtSecret string, jwtExpiresInHours int) *AuthService {
	return &AuthService{
		userRepo:          userRepo,
		jwtSecret:         jwtSecret,
		jwtExpiresInHours: jwtExpiresInHours,
	}
}

// Register validates user credentials, ensures email uniqueness, hashes the password, and creates the account.
func (s *AuthService) Register(ctx context.Context, email, password string) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || len(password) < 6 {
		return nil, shared.ErrBadRequest
	}

	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, shared.ErrConflict
	}
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return nil, shared.ErrInternalServer
	}

	hashedPassword, err := shared.HashPassword(password)
	if err != nil {
		return nil, shared.ErrInternalServer
	}

	user := &User{
		Email:    email,
		Password: hashedPassword,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, shared.ErrInternalServer
	}

	return user, nil
}

// Login validates user email and password, returning the user and a signed JWT token.
func (s *AuthService) Login(ctx context.Context, email, password string) (*User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return nil, "", shared.ErrUnauthorized
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, "", shared.ErrUnauthorized
		}
		return nil, "", shared.ErrInternalServer
	}

	if !shared.CheckPassword(user.Password, password) {
		return nil, "", shared.ErrUnauthorized
	}

	token, err := shared.GenerateToken(user.ID, user.Email, s.jwtSecret, s.jwtExpiresInHours)
	if err != nil {
		return nil, "", shared.ErrInternalServer
	}

	return user, token, nil
}
