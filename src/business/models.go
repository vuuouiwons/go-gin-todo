package business

import (
	"time"

	"go-gin-todo/src/shared"
)

// User represents a user account domain entity.
type User struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Todolist represents a grouped collection of tasks domain entity.
type Todolist struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Todo represents an individual actionable task item domain entity.
type Todo struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	Message   string    `json:"message"`
	Check     bool      `json:"check"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TodoUpdateItem carries update fields for a specific Todo entity.
type TodoUpdateItem struct {
	ID      uint
	Message *string
	Check   *bool
}

// Pagination aliases shared.Pagination for business responses.
type Pagination = shared.Pagination
