package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go-gin-todo/internal/app"
)

// AuthHandler handles HTTP requests for user registration and login.
type AuthHandler struct {
	commands *app.AuthCommands
}

// NewAuthHandler constructs an AuthHandler backed by AuthCommands.
func NewAuthHandler(commands *app.AuthCommands) *AuthHandler {
	return &AuthHandler{commands: commands}
}

// Register handles POST /register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.commands.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, app.ErrConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
			return
		}
		respondDomainError(c, err, "", "Invalid registration details", "Failed to register user")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"data":    ToUserResponse(user),
	})
}

// Login handles POST /login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.commands.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, app.ErrUnauthorized) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		respondDomainError(c, err, "", "", "Failed to authenticate user")
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Message: "Login successful",
		Token:   token,
		Data:    ToUserResponse(user),
	})
}
