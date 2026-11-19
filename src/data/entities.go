package data

import (
	"time"

	"go-gin-todo/src/business"
	"gorm.io/gorm"
)

// UserEntity represents the database schema for users.
type UserEntity struct {
	ID        uint           `gorm:"primaryKey"`
	Email     string         `gorm:"uniqueIndex;not null"`
	Password  string         `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (UserEntity) TableName() string {
	return "users"
}

func UserToEntity(u *business.User) *UserEntity {
	if u == nil {
		return nil
	}
	return &UserEntity{
		ID:        u.ID,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func UserToDomain(e *UserEntity) *business.User {
	if e == nil {
		return nil
	}
	return &business.User{
		ID:        e.ID,
		Email:     e.Email,
		Password:  e.Password,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

// TodolistEntity represents the database schema for todolists.
type TodolistEntity struct {
	ID        uint           `gorm:"primaryKey"`
	UserID    uint           `gorm:"not null;index"`
	Title     string         `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	User      UserEntity     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}

func (TodolistEntity) TableName() string {
	return "todolists"
}

func TodolistToEntity(t *business.Todolist) *TodolistEntity {
	if t == nil {
		return nil
	}
	return &TodolistEntity{
		ID:        t.ID,
		UserID:    t.UserID,
		Title:     t.Title,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func TodolistToDomain(e *TodolistEntity) *business.Todolist {
	if e == nil {
		return nil
	}
	return &business.Todolist{
		ID:        e.ID,
		UserID:    e.UserID,
		Title:     e.Title,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func TodolistListToDomain(entities []TodolistEntity) []business.Todolist {
	res := make([]business.Todolist, len(entities))
	for i := range entities {
		res[i] = *TodolistToDomain(&entities[i])
	}
	return res
}

// TodoEntity represents the database schema for todos.
type TodoEntity struct {
	ID        uint           `gorm:"primaryKey"`
	UserID    uint           `gorm:"not null;index"`
	Message   string         `gorm:"not null"`
	Check     bool           `gorm:"not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	User      UserEntity     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}

func (TodoEntity) TableName() string {
	return "todos"
}

func TodoToEntity(t *business.Todo) *TodoEntity {
	if t == nil {
		return nil
	}
	return &TodoEntity{
		ID:        t.ID,
		UserID:    t.UserID,
		Message:   t.Message,
		Check:     t.Check,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func TodoToDomain(e *TodoEntity) *business.Todo {
	if e == nil {
		return nil
	}
	return &business.Todo{
		ID:        e.ID,
		UserID:    e.UserID,
		Message:   e.Message,
		Check:     e.Check,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func TodoListToDomain(entities []TodoEntity) []business.Todo {
	res := make([]business.Todo, len(entities))
	for i := range entities {
		res[i] = *TodoToDomain(&entities[i])
	}
	return res
}
