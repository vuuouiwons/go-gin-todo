package domain

import (
	"context"
	"time"
)

// Todo represents the Todo table: Todo (id, user_id, message, check)
type Todo struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"column:user_id;index;not null" json:"userId"`
	Message   string    `gorm:"column:message;type:text;not null" json:"message"`
	Check     bool      `gorm:"column:check;default:false" json:"check"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Todo) TableName() string {
	return "todos"
}

// TodoUpdateItem carries update fields for a todo
type TodoUpdateItem struct {
	ID      uint
	Message *string
	Check   *bool
}

// TodoRepository defines the persistence contract for Todo entities
type TodoRepository interface {
	Create(ctx context.Context, todo *Todo) error
	CreateBulk(ctx context.Context, todos []Todo) ([]Todo, error)
	GetByID(ctx context.Context, id uint, userID uint) (*Todo, error)
	GetBulk(ctx context.Context, userID uint, search string, checkFilter *bool, page, pageSize int) ([]Todo, int64, error)
	Update(ctx context.Context, todo *Todo) error
	UpdateBulk(ctx context.Context, userID uint, updates []TodoUpdateItem) ([]Todo, error)
	DeleteByID(ctx context.Context, id uint, userID uint) (int64, error)
	DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error)
}
