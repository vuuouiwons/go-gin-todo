package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-gin-todo/internal/app"
)

const (
	defaultPage  = 1
	defaultLimit = 10
)

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

func parseIDParam(c *gin.Context, entityName string) (uint, bool) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid %s ID", entityName)})
		return 0, false
	}
	return uint(id), true
}

func readTrimmedBody(c *gin.Context, emptyErrMsg string) ([]byte, bool) {
	bodyBytes, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": emptyErrMsg})
		return nil, false
	}

	trimmed := bytes.TrimSpace(bodyBytes)
	if len(trimmed) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": emptyErrMsg})
		return nil, false
	}

	return trimmed, true
}

func parseDeleteIDs(c *gin.Context) ([]uint, bool) {
	trimmed, ok := readTrimmedBody(c, "Request body cannot be empty. Deletion IDs must be provided in the body (e.g. [1, 2] or {'id': 1} or {'ids': [1, 2]})")
	if !ok {
		return nil, false
	}

	var idsToDelete []uint

	if trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &idsToDelete); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array of IDs: " + err.Error()})
			return nil, false
		}
	} else {
		var payload DeleteRequest
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

func respondDomainError(c *gin.Context, err error, notFoundMsg, badReqMsg, defaultErrMsg string) {
	if errors.Is(err, app.ErrNotFound) {
		msg := notFoundMsg
		if msg == "" {
			msg = "Resource not found"
		}
		c.JSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}
	if errors.Is(err, app.ErrBadRequest) {
		msg := badReqMsg
		if msg == "" {
			msg = "Invalid request data"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	if errors.Is(err, app.ErrUnauthorized) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if errors.Is(err, app.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{"error": "Resource conflict"})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": defaultErrMsg})
}
