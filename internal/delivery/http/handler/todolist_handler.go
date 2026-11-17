package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
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
	page, limit := parsePagination(c)

	todolists, pagination, err := h.usecase.GetBulk(c.Request.Context(), userID, search, page, limit)
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

	id, ok := parseIDParam(c, "todolist")
	if !ok {
		return
	}

	todolist, err := h.usecase.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		respondDomainError(c, err, "Todolist not found", "", "Failed to fetch todolist")
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

	// Bulk Array format: [{"title": "..."}, ...]
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
			respondDomainError(c, err, "", "Title cannot be empty for any item", "Failed to create todolists")
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Todolists created successfully",
			"count":   len(created),
			"data":    created,
		})
		return
	}

	// Object format: single {"title": "..."} or wrapped bulk {"items": [...]}
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
			respondDomainError(c, err, "", "Title cannot be empty for any item", "Failed to create todolists")
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
		respondDomainError(c, err, "", "Field 'title' is required", "Failed to create todolist")
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
		respondDomainError(c, err, "One or more todolists not found or unauthorized", "Valid 'id' and 'title' are required for each item", "Failed to update todolists")
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

	id, ok := parseIDParam(c, "todolist")
	if !ok {
		return
	}

	var req dto.CreateTodolistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.usecase.UpdateByID(c.Request.Context(), id, userID, req.Title)
	if err != nil {
		respondDomainError(c, err, "Todolist not found", "Field 'title' cannot be empty", "Failed to update todolist")
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

	idsToDelete, ok := parseDeleteIDs(c)
	if !ok {
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

	id, ok := parseIDParam(c, "todolist")
	if !ok {
		return
	}

	deletedCount, err := h.usecase.DeleteByID(c.Request.Context(), id, userID)
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
