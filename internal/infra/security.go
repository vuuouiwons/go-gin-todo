package infra

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go-gin-todo/internal/app"
	"golang.org/x/crypto/bcrypt"
)

type bcryptHasher struct {
	cost int
}

// NewBcryptHasher constructs a bcrypt implementation of app.PasswordHasher.
func NewBcryptHasher(cost int) app.PasswordHasher {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	return &bcryptHasher{cost: cost}
}

func (b *bcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	return string(bytes), err
}

func (b *bcryptHasher) Compare(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

type jwtCustomClaims struct {
	UserID uint   `json:"userId"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type jwtTokenManager struct {
	secret         []byte
	expireDuration time.Duration
	issuer         string
}

// NewJWTTokenManager constructs a JWT implementation of app.TokenManager.
func NewJWTTokenManager(secret string, expireHours int) app.TokenManager {
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

func (j *jwtTokenManager) Validate(tokenStr string) (*app.TokenClaims, error) {
	claims := &jwtCustomClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})

	if err != nil || !token.Valid {
		return nil, app.ErrUnauthorized
	}

	return &app.TokenClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
	}, nil
}
