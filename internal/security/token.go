package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go-gin-todo/internal/domain"
)

type TokenClaims struct {
	UserID uint   `json:"userId"`
	Email  string `json:"email"`
}

type jwtCustomClaims struct {
	UserID uint   `json:"userId"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type TokenManager interface {
	Generate(userID uint, email string) (string, error)
	Validate(tokenStr string) (*TokenClaims, error)
}

type jwtTokenManager struct {
	secret        []byte
	expireDuration time.Duration
	issuer        string
}

func NewJWTTokenManager(secret string, expireHours int) TokenManager {
	if expireHours <= 0 {
		expireHours = 24
	}
	return &jwtTokenManager{
		secret:         []byte(secret),
		expireDuration: time.Duration(expireHours) * time.Hour,
		issuer:         "go-gin-todo-api",
	}
}

func (j *jwtTokenManager) Generate(userID uint, email string) (string, error) {
	now := time.Now()
	claims := jwtCustomClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expireDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    j.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *jwtTokenManager) Validate(tokenStr string) (*TokenClaims, error) {
	claims := &jwtCustomClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})

	if err != nil || !token.Valid {
		return nil, domain.ErrUnauthorized
	}

	return &TokenClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
	}, nil
}
