package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go-gin-todo/internal/security"
)

const (
	ContextUserIDKey    = "userID"
	ContextUserEmailKey = "userEmail"
)

func AuthMiddleware(tokenManager security.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header must be in format: Bearer <token>",
			})
			return
		}

		tokenString := parts[1]
		claims, err := tokenManager.Validate(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUserEmailKey, claims.Email)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (uint, error) {
	val, exists := c.Get(ContextUserIDKey)
	if !exists {
		return 0, errors.New("user ID not found in context")
	}

	userID, ok := val.(uint)
	if !ok {
		return 0, errors.New("invalid user ID type in context")
	}

	return userID, nil
}
