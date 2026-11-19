package presentation

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-gin-todo/src/business"
)

// AuthController handles HTTP requests for user authentication.
type AuthController struct {
	authService *business.AuthService
}

// NewAuthController constructs a new AuthController.
func NewAuthController(authService *business.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// Register handles POST /register.
func (h *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		respondDomainError(c, err, "", "Password must be at least 6 characters", "Failed to register user")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    ToUserResponse(user),
	})
}

// Login handles POST /login.
func (h *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	user, token, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		respondDomainError(c, err, "", "", "Failed to login")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user":    ToUserResponse(user),
	})
}
