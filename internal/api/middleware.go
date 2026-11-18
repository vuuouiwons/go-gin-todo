package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go-gin-todo/internal/app"
)

var errTransactionRollback = errors.New("rollback transaction due to HTTP error response")

const (
	ContextUserIDKey    = "userID"
	ContextUserEmailKey = "userEmail"
)

// AuthMiddleware validates Bearer JWT tokens using app.TokenManager.
func AuthMiddleware(tokenManager app.TokenManager) gin.HandlerFunc {
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

		claims, err := tokenManager.Validate(parts[1])
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

// GetUserID retrieves the authenticated user's ID from the Gin context.
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

// TransactionMiddleware wraps mutating HTTP requests in an atomic database transaction
// via context propagation (similar to NestJS's @Transactional() CLS decorator).
// Commits on status < 400 with no Gin errors; rolls back on status >= 400, errors, or panic.
func TransactionMiddleware(transactor app.Transactor) gin.HandlerFunc {
	return func(c *gin.Context) {
		origReq := c.Request
		err := transactor.ExecInTransaction(origReq.Context(), func(txCtx context.Context) error {
			c.Request = origReq.WithContext(txCtx)
			c.Next()

			if c.Writer.Status() >= http.StatusBadRequest || len(c.Errors) > 0 {
				return errTransactionRollback
			}
			return nil
		})
		c.Request = origReq
		if err != nil && !errors.Is(err, errTransactionRollback) && !c.Writer.Written() {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		}
	}
}

// CORSMiddleware configures permissive CORS headers for API access.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
