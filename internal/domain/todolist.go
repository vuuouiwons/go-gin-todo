package domain

import (
	"context"
	"time"
)

// Todolist represents the Todolist table: Todolist (id, userId, title)
type Todolist struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"column:userId;index;not null" json:"userId"`
	Title     string    `gorm:"column:title;type:varchar(255);not null" json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Todolist) TableName() string {
	return "todolists"
}

// TodolistRepository defines the persistence contract for Todolist entities
type TodolistRepository interface {
	Create(ctx context.Context, todolist *Todolist) error
	CreateBulk(ctx context.Context, todolists []Todolist) ([]Todolist, error)
	GetByID(ctx context.Context, id uint, userID uint) (*Todolist, error)
	GetBulk(ctx context.Context, userID uint, search string, page, pageSize int) ([]Todolist, int64, error)
	Update(ctx context.Context, todolist *Todolist) error
	UpdateBulk(ctx context.Context, userID uint, updates []Todolist) ([]Todolist, error)
	DeleteByID(ctx context.Context, id uint, userID uint) (int64, error)
	DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error)
}
