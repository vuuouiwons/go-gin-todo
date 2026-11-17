package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go-gin-todo/internal/delivery/http/dto"
	"go-gin-todo/internal/domain"
	"go-gin-todo/internal/usecase"
)

type AuthHandler struct {
	usecase usecase.AuthUsecase
}

func NewAuthHandler(u usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{usecase: u}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.usecase.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
			return
		}
		if errors.Is(err, domain.ErrBadRequest) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid registration details"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"data": dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.usecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate user"})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		Message: "Login successful",
		Token:   token,
		Data: dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	})
}
