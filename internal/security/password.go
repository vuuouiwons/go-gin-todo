package security

import (
	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) bool
}

type bcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) PasswordHasher {
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
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
