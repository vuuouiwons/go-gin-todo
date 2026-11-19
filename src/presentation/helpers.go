package presentation

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-gin-todo/src/shared"
)

// readTrimmedBody reads the request body and returns trimmed bytes while preserving body for future reads.
func readTrimmedBody(c *gin.Context, emptyErrMsg string) ([]byte, bool) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil || len(bodyBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": emptyErrMsg})
		return nil, false
	}

	trimmed := bytes.TrimSpace(bodyBytes)
	if len(trimmed) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": emptyErrMsg})
		return nil, false
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return trimmed, true
}

// parseIDParam extracts and validates an unsigned integer route parameter.
func parseIDParam(c *gin.Context, resourceName string) (uint, bool) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid " + resourceName + " ID"})
		return 0, false
	}
	return uint(id), true
}

// respondDomainError maps application errors to appropriate HTTP status codes.
func respondDomainError(c *gin.Context, err error, notFoundMsg, badReqMsg, defaultMsg string) {
	if errors.Is(err, shared.ErrNotFound) {
		msg := notFoundMsg
		if msg == "" {
			msg = "Resource not found"
		}
		c.JSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	if errors.Is(err, shared.ErrBadRequest) {
		msg := badReqMsg
		if msg == "" {
			msg = "Invalid request payload"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	if errors.Is(err, shared.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	if errors.Is(err, shared.ErrUnauthorized) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	msg := defaultMsg
	if msg == "" {
		msg = "Internal server error"
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
}
