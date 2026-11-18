package app

import (
	"errors"
	"math"
	"time"
)

// Application domain errors.
var (
	ErrNotFound       = errors.New("resource not found")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrConflict       = errors.New("resource already exists")
	ErrBadRequest     = errors.New("invalid input data")
	ErrInternalServer = errors.New("internal server error")
)

// User represents a registered user account in the users table.
type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Password  string    `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (User) TableName() string {
	return "users"
}

// Todolist represents a todolist owned by a user in the todolists table.
type Todolist struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"column:user_id;index;not null"`
	Title     string    `gorm:"column:title;type:varchar(255);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Todolist) TableName() string {
	return "todolists"
}

// Todo represents a task item owned by a user in the todos table.
type Todo struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"column:user_id;index;not null"`
	Message   string    `gorm:"column:message;type:text;not null"`
	Check     bool      `gorm:"column:check;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Todo) TableName() string {
	return "todos"
}

// TodoUpdateItem carries partial update fields for a Todo item.
type TodoUpdateItem struct {
	ID      uint
	Message *string
	Check   *bool
}

// TokenClaims represents authenticated user identity extracted from a token.
type TokenClaims struct {
	UserID uint
	Email  string
}

// Pagination holds pagination metadata computed from query results.
type Pagination struct {
	Page       int
	Limit      int
	TotalItems int64
	TotalPages int
	HasNext    bool
	HasPrev    bool
}

// NewPagination constructs a Pagination struct with computed metadata.
func NewPagination(page, limit int, totalItems int64) Pagination {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	var totalPages int
	if totalItems > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(limit)))
	}

	return Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1 && (totalPages == 0 || page <= totalPages+1),
	}
}
