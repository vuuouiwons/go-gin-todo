package api

import (
	"time"

	"go-gin-todo/internal/app"
)

// Request DTOs

type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type CreateTodolistRequest struct {
	Title string `json:"title" binding:"required"`
}

type UpdateTodolistRequest struct {
	ID    uint   `json:"id"`
	Title string `json:"title" binding:"required"`
}

type CreateTodoRequest struct {
	Message string `json:"message" binding:"required"`
	Check   *bool  `json:"check"`
}

type UpdateTodoRequest struct {
	ID      uint    `json:"id"`
	Message *string `json:"message"`
	Check   *bool   `json:"check"`
}

type DeleteRequest struct {
	ID  uint   `json:"id"`
	IDs []uint `json:"ids"`
}

// Response DTOs

type UserResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type LoginResponse struct {
	Message string       `json:"message"`
	Token   string       `json:"token"`
	Data    UserResponse `json:"data"`
}

type TodolistResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TodoResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	Message   string    `json:"message"`
	Check     bool      `json:"check"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}

// Explicit API Mappers (Security Boundary)

func ToUserResponse(u *app.User) UserResponse {
	if u == nil {
		return UserResponse{}
	}
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

func ToTodolistResponse(t *app.Todolist) TodolistResponse {
	if t == nil {
		return TodolistResponse{}
	}
	return TodolistResponse{
		ID:        t.ID,
		UserID:    t.UserID,
		Title:     t.Title,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func ToTodolistResponseList(items []app.Todolist) []TodolistResponse {
	res := make([]TodolistResponse, 0, len(items))
	for i := range items {
		res = append(res, ToTodolistResponse(&items[i]))
	}
	return res
}

func ToTodoResponse(t *app.Todo) TodoResponse {
	if t == nil {
		return TodoResponse{}
	}
	return TodoResponse{
		ID:        t.ID,
		UserID:    t.UserID,
		Message:   t.Message,
		Check:     t.Check,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func ToTodoResponseList(items []app.Todo) []TodoResponse {
	res := make([]TodoResponse, 0, len(items))
	for i := range items {
		res = append(res, ToTodoResponse(&items[i]))
	}
	return res
}

func ToPaginationResponse(p app.Pagination) PaginationResponse {
	return PaginationResponse{
		Page:       p.Page,
		Limit:      p.Limit,
		TotalItems: p.TotalItems,
		TotalPages: p.TotalPages,
		HasNext:    p.HasNext,
		HasPrev:    p.HasPrev,
	}
}
