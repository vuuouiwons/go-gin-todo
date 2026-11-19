package presentation

import (
	"time"

	"go-gin-todo/src/business"
	"go-gin-todo/src/shared"
)

// --- Request DTOs ---

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateTodoRequest struct {
	Message string `json:"message"`
	Check   *bool  `json:"check"`
}

type UpdateTodoRequest struct {
	ID      uint    `json:"id"`
	Message *string `json:"message"`
	Check   *bool   `json:"check"`
}

type CreateTodolistRequest struct {
	Title string `json:"title"`
}

type UpdateTodolistRequest struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
}

type DeleteRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

type DeleteBulkRequest = DeleteRequest

// --- Outbound Response Payloads (Boundary Security) ---

type UserResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

func ToUserResponse(u *business.User) UserResponse {
	return UserResponse{
		ID:    u.ID,
		Email: u.Email,
	}
}

type TodoResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	Message   string    `json:"message"`
	Check     bool      `json:"check"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func ToTodoResponse(t *business.Todo) TodoResponse {
	return TodoResponse{
		ID:        t.ID,
		UserID:    t.UserID,
		Message:   t.Message,
		Check:     t.Check,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func ToTodoResponseList(todos []business.Todo) []TodoResponse {
	res := make([]TodoResponse, len(todos))
	for i := range todos {
		res[i] = ToTodoResponse(&todos[i])
	}
	return res
}

type TodolistResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func ToTodolistResponse(tl *business.Todolist) TodolistResponse {
	return TodolistResponse{
		ID:        tl.ID,
		UserID:    tl.UserID,
		Title:     tl.Title,
		CreatedAt: tl.CreatedAt,
		UpdatedAt: tl.UpdatedAt,
	}
}

func ToTodolistResponseList(todolists []business.Todolist) []TodolistResponse {
	res := make([]TodolistResponse, len(todolists))
	for i := range todolists {
		res[i] = ToTodolistResponse(&todolists[i])
	}
	return res
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}

func ToPaginationResponse(p shared.Pagination) PaginationResponse {
	return PaginationResponse{
		Page:       p.Page,
		Limit:      p.Limit,
		TotalItems: p.TotalItems,
		TotalPages: p.TotalPages,
		HasNext:    p.HasNext,
		HasPrev:    p.HasPrev,
	}
}
