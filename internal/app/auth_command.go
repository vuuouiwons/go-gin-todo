package app

import (
	"context"
	"errors"
	"strings"
)

// AuthCommands handles authentication write operations and credential checks.
type AuthCommands struct {
	userRepo     UserRepository
	hasher       PasswordHasher
	tokenManager TokenManager
	tx           Transactor
}

// NewAuthCommands constructs a new AuthCommands handler.
func NewAuthCommands(userRepo UserRepository, hasher PasswordHasher, tokenManager TokenManager, tx Transactor) *AuthCommands {
	return &AuthCommands{
		userRepo:     userRepo,
		hasher:       hasher,
		tokenManager: tokenManager,
		tx:           tx,
	}
}

// Register creates a new user account with a bcrypt-hashed password inside a transaction.
func (c *AuthCommands) Register(ctx context.Context, email, password string) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || len(password) < 6 {
		return nil, ErrBadRequest
	}

	hashedPassword, err := c.hasher.Hash(password)
	if err != nil {
		return nil, ErrInternalServer
	}

	return RunInTx(ctx, c.tx, func(txCtx context.Context) (*User, error) {
		existing, err := c.userRepo.GetByEmail(txCtx, email)
		if err == nil && existing != nil {
			return nil, ErrConflict
		}
		if err != nil && !errors.Is(err, ErrNotFound) {
			return nil, ErrInternalServer
		}

		newUser := &User{
			Email:    email,
			Password: hashedPassword,
		}

		if err := c.userRepo.Create(txCtx, newUser); err != nil {
			return nil, ErrInternalServer
		}

		return newUser, nil
	})
}

// Login validates credentials and returns the user and signed JWT token.
func (c *AuthCommands) Login(ctx context.Context, email, password string) (*User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return nil, "", ErrUnauthorized
	}

	user, err := c.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, "", ErrUnauthorized
	}

	if !c.hasher.Compare(user.Password, password) {
		return nil, "", ErrUnauthorized
	}

	token, err := c.tokenManager.Generate(user.ID, user.Email)
	if err != nil {
		return nil, "", ErrInternalServer
	}

	return user, token, nil
}
