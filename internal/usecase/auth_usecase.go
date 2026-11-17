package usecase

import (
	"context"
	"strings"

	"go-gin-todo/internal/domain"
	"go-gin-todo/internal/security"
)

type AuthUsecase interface {
	Register(ctx context.Context, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.User, string, error)
}

type authUsecase struct {
	userRepo     domain.UserRepository
	hasher       security.PasswordHasher
	tokenManager security.TokenManager
}

func NewAuthUsecase(
	userRepo domain.UserRepository,
	hasher security.PasswordHasher,
	tokenManager security.TokenManager,
) AuthUsecase {
	return &authUsecase{
		userRepo:     userRepo,
		hasher:       hasher,
		tokenManager: tokenManager,
	}
}

func (u *authUsecase) Register(ctx context.Context, email, password string) (*domain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || len(password) < 6 {
		return nil, domain.ErrBadRequest
	}

	// Check if user already exists
	existing, err := u.userRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, domain.ErrConflict
	}

	hashedPassword, err := u.hasher.Hash(password)
	if err != nil {
		return nil, domain.ErrInternalServer
	}

	newUser := &domain.User{
		Email:    email,
		Password: hashedPassword,
	}

	if err := u.userRepo.Create(ctx, newUser); err != nil {
		return nil, domain.ErrInternalServer
	}

	return newUser, nil
}

func (u *authUsecase) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return nil, "", domain.ErrUnauthorized
	}

	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, "", domain.ErrUnauthorized
	}

	if !u.hasher.Compare(user.Password, password) {
		return nil, "", domain.ErrUnauthorized
	}

	token, err := u.tokenManager.Generate(user.ID, user.Email)
	if err != nil {
		return nil, "", domain.ErrInternalServer
	}

	return user, token, nil
}
