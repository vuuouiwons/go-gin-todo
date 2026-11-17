package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-gin-todo/internal/delivery/http/dto"
	"go-gin-todo/internal/domain"
)

const (
	defaultPage  = 1
	defaultLimit = 10
)

// parsePagination extracts and validates page and limit query parameters.
func parsePagination(c *gin.Context) (int, int) {
	page, err := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(defaultPage)))
	if err != nil || page <= 0 {
		page = defaultPage
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if err != nil || limit <= 0 {
		limit = defaultLimit
	}

	return page, limit
}

// parseIDParam extracts and validates the :id route parameter.
func parseIDParam(c *gin.Context, entityName string) (uint, bool) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid %s ID", entityName)})
		return 0, false
	}
	return uint(id), true
}

// parseDeleteIDs extracts single or bulk deletion IDs strictly from the request body.
// Supports both JSON array [1, 2] and JSON object {"id": 1} or {"ids": [1, 2]}.
func parseDeleteIDs(c *gin.Context) ([]uint, bool) {
	bodyBytes, err := c.GetRawData()
	if err != nil || len(bodyBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request body cannot be empty. Deletion IDs must be provided in the body (e.g. [1, 2] or {'id': 1} or {'ids': [1, 2]})",
		})
		return nil, false
	}

	trimmed := bytes.TrimSpace(bodyBytes)
	if len(trimmed) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request body cannot be empty. Deletion IDs must be provided in the body (e.g. [1, 2] or {'id': 1} or {'ids': [1, 2]})",
		})
		return nil, false
	}

	var idsToDelete []uint

	if trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &idsToDelete); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array of IDs: " + err.Error()})
			return nil, false
		}
	} else {
		var payload dto.DeleteRequest
		if err := json.Unmarshal(trimmed, &payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON object format: " + err.Error()})
			return nil, false
		}
		if len(payload.IDs) > 0 {
			idsToDelete = payload.IDs
		} else if payload.ID > 0 {
			idsToDelete = append(idsToDelete, payload.ID)
		}
	}

	if len(idsToDelete) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No ID or IDs specified in body. Provide [1, 2] or {'id': 1} or {'ids': [1, 2]}",
		})
		return nil, false
	}

	return idsToDelete, true
}

// respondDomainError maps domain errors to appropriate HTTP status codes and responses.
func respondDomainError(c *gin.Context, err error, notFoundMsg, badReqMsg, defaultErrMsg string) {
	if errors.Is(err, domain.ErrNotFound) {
		msg := notFoundMsg
		if msg == "" {
			msg = "Resource not found"
		}
		c.JSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}
	if errors.Is(err, domain.ErrBadRequest) {
		msg := badReqMsg
		if msg == "" {
			msg = "Invalid request data"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	if errors.Is(err, domain.ErrUnauthorized) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if errors.Is(err, domain.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{"error": "Resource conflict"})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": defaultErrMsg})
}
