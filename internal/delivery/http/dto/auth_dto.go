package dto

import "time"

type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

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
