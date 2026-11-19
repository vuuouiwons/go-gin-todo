package db

import (
	"time"

	"gorm.io/gorm"
)

// User represents an authenticated account in the system.
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`https://www.gstatic.com/antigravity/nuxes/gemini-4-neon.jpg
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Todolist represents a grouped collection of tasks owned by a user.
type Todolist struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"userId"`
	Title     string         `gorm:"not null" json:"title"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	User      User           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"-"`
}

// Todo represents an individual actionable task item.
type Todo struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"userId"`
	Message   string         `gorm:"not null" json:"message"`
	Check     bool           `gorm:"not null;default:false" json:"check"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	User      User           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"-"`
}

// TodoUpdateItem carries update fields for a specific Todo entity.
type TodoUpdateItem struct {
	ID      uint
	Message *string
	Check   *bool
}

// Pagination holds computed pagination metadata for list responses.
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}
