package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go-gin-todo/internal/delivery/http/dto"
	"go-gin-todo/internal/delivery/http/middleware"
	"go-gin-todo/internal/domain"
	"go-gin-todo/internal/usecase"
)

type TodolistHandler struct {
	usecase usecase.TodolistUsecase
}

func NewTodolistHandler(u usecase.TodolistUsecase) *TodolistHandler {
	return &TodolistHandler{usecase: u}
}

// GetBulk handles GET /todolist (bulk) with pagination
func (h *TodolistHandler) GetBulk(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	search := c.Query("search")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	// Support pageSize aliases like limit or per_page for convenience
	if c.Query("pageSize") == "" {
		if c.Query("limit") != "" {
			pageSizeStr = c.Query("limit")
		} else if c.Query("per_page") != "" {
			pageSizeStr = c.Query("per_page")
		}
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	todolists, pagination, err := h.usecase.GetBulk(c.Request.Context(), userID, search, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch todolists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       todolists,
		"pagination": pagination,
	})
}

// GetOne handles GET /todolist/:id
func (h *TodolistHandler) GetOne(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todolist ID"})
		return
	}

	todolist, err := h.usecase.GetByID(c.Request.Context(), uint(id), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todolist not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch todolist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": todolist})
}

// Create handles POST /todolist (one, bulk)
func (h *TodolistHandler) Create(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	bodyBytes, err := c.GetRawData()
	if err != nil || len(bodyBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body cannot be empty"})
		return
	}

	trimmed := bytes.TrimSpace(bodyBytes)
	if len(trimmed) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body cannot be empty"})
		return
	}

	// Bulk Array
	if trimmed[0] == '[' {
		var inputs []dto.CreateTodolistRequest
		if err := json.Unmarshal(trimmed, &inputs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array format: " + err.Error()})
			return
		}

		var titles []string
		for _, item := range inputs {
			titles = append(titles, item.Title)
		}

		created, err := h.usecase.CreateBulk(c.Request.Context(), userID, titles)
		if err != nil {
			if errors.Is(err, domain.ErrBadRequest) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty for any item"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todolists"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Todolists created successfully",
			"count":   len(created),
			"data":    created,
		})
		return
	}

	// Object (single or wrapped bulk)
	var wrapped struct {
		Items []dto.CreateTodolistRequest `json:"items"`
		Title string                      `json:"title"`
	}
	if err := json.Unmarshal(trimmed, &wrapped); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON object format: " + err.Error()})
		return
	}

	if len(wrapped.Items) > 0 {
		var titles []string
		for _, item := range wrapped.Items {
			titles = append(titles, item.Title)
		}

		created, err := h.usecase.CreateBulk(c.Request.Context(), userID, titles)
		if err != nil {
			if errors.Is(err, domain.ErrBadRequest) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty for any item"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todolists"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Todolists created successfully",
			"count":   len(created),
			"data":    created,
		})
		return
	}

	// Single item
	created, err := h.usecase.CreateOne(c.Request.Context(), userID, wrapped.Title)
	if err != nil {
		if errors.Is(err, domain.ErrBadRequest) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Field 'title' is required"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todolist"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Todolist created successfully",
		"data":    created,
	})
}

// PatchOneOrBulk handles PATCH /todolist (update by id, one, bulk)
func (h *TodolistHandler) PatchOneOrBulk(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	bodyBytes, err := c.GetRawData()
	if err != nil || len(bodyBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body cannot be empty"})
		return
	}

	trimmed := bytes.TrimSpace(bodyBytes)
	if len(trimmed) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body cannot be empty"})
		return
	}

	var itemsToUpdate []dto.UpdateTodolistRequest

	if trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &itemsToUpdate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array format: " + err.Error()})
			return
		}
	} else {
		var wrapped struct {
			Items []dto.UpdateTodolistRequest `json:"items"`
			ID    uint                        `json:"id"`
			Title string                      `json:"title"`
		}
		if err := json.Unmarshal(trimmed, &wrapped); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON object format: " + err.Error()})
			return
		}

		if len(wrapped.Items) > 0 {
			itemsToUpdate = wrapped.Items
		} else {
			if wrapped.ID == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Field 'id' is required for update"})
				return
			}
			itemsToUpdate = []dto.UpdateTodolistRequest{{
				ID:    wrapped.ID,
				Title: wrapped.Title,
			}}
		}
	}

	if len(itemsToUpdate) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No items provided for update"})
		return
	}

	var domainUpdates []domain.Todolist
	for _, item := range itemsToUpdate {
		domainUpdates = append(domainUpdates, domain.Todolist{
			ID:    item.ID,
			Title: strings.TrimSpace(item.Title),
		})
	}

	updated, err := h.usecase.UpdateBulk(c.Request.Context(), userID, domainUpdates)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "One or more todolists not found or unauthorized"})
			return
		}
		if errors.Is(err, domain.ErrBadRequest) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Valid 'id' and 'title' are required for each item"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update todolists"})
		return
	}

	if len(updated) == 1 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Todolist updated successfully",
			"data":    updated[0],
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Todolists updated successfully",
			"count":   len(updated),
			"data":    updated,
		})
	}
}

// PatchByID handles PATCH /todolist/:id
func (h *TodolistHandler) PatchByID(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todolist ID"})
		return
	}

	var req dto.CreateTodolistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.usecase.UpdateByID(c.Request.Context(), uint(id), userID, req.Title)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todolist not found"})
			return
		}
		if errors.Is(err, domain.ErrBadRequest) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Field 'title' cannot be empty"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update todolist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Todolist updated successfully",
		"data":    updated,
	})
}

// DeleteOneOrBulk handles DELETE /todolist (one, bulk)
// Deletion IDs must strictly be provided in the request body.
func (h *TodolistHandler) DeleteOneOrBulk(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	bodyBytes, err := c.GetRawData()
	if err != nil || len(bodyBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body cannot be empty. Deletion IDs must be provided in the body (e.g. [1, 2] or {'id': 1} or {'ids': [1, 2]})"})
		return
	}

	trimmed := bytes.TrimSpace(bodyBytes)
	if len(trimmed) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body cannot be empty. Deletion IDs must be provided in the body (e.g. [1, 2] or {'id': 1} or {'ids': [1, 2]})"})
		return
	}

	var idsToDelete []uint

	if trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &idsToDelete); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array of IDs: " + err.Error()})
			return
		}
	} else {
		var payload dto.DeleteTodolistRequest
		if err := json.Unmarshal(trimmed, &payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON object format: " + err.Error()})
			return
		}
		if len(payload.IDs) > 0 {
			idsToDelete = payload.IDs
		} else if payload.ID > 0 {
			idsToDelete = append(idsToDelete, payload.ID)
		}
	}

	if len(idsToDelete) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No ID or IDs specified in body. Provide [1, 2] or {'id': 1} or {'ids': [1, 2]}"})
		return
	}

	deletedCount, err := h.usecase.DeleteBulk(c.Request.Context(), idsToDelete, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todolists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Todolist(s) deleted successfully",
		"deletedCount": deletedCount,
		"requestedIDs": idsToDelete,
	})
}

// DeleteByID handles DELETE /todolist/:id
func (h *TodolistHandler) DeleteByID(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todolist ID"})
		return
	}

	deletedCount, err := h.usecase.DeleteByID(c.Request.Context(), uint(id), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todolist"})
		return
	}

	if deletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todolist not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Todolist deleted successfully",
		"id":      id,
	})
}
