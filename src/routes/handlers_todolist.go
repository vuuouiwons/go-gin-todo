package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"go-gin-todo/src/db"
	"go-gin-todo/src/services"
	"go-gin-todo/src/utils"
)

// TodolistHandler handles HTTP requests for Todolist resources.
type TodolistHandler struct {
	service *services.TodolistService
}

// NewTodolistHandler constructs a new TodolistHandler.
func NewTodolistHandler(service *services.TodolistService) *TodolistHandler {
	return &TodolistHandler{service: service}
}

// GetBulk handles GET /todolist.
func (h *TodolistHandler) GetBulk(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	search := c.Query("search")
	page, limit := utils.ParsePagination(c)

	items, pagination, err := h.service.GetBulk(c.Request.Context(), userID, search, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch todolists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       ToTodolistResponseList(items),
		"pagination": ToPaginationResponse(pagination),
	})
}

// GetOne handles GET /todolist/:id.
func (h *TodolistHandler) GetOne(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id, ok := parseIDParam(c, "todolist")
	if !ok {
		return
	}

	item, err := h.service.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		respondDomainError(c, err, "Todolist not found", "", "Failed to fetch todolist")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ToTodolistResponse(item)})
}

// Create handles POST /todolist (single or bulk).
func (h *TodolistHandler) Create(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	trimmed, ok := readTrimmedBody(c, "Request body cannot be empty")
	if !ok {
		return
	}

	if trimmed[0] == '[' {
		var inputs []CreateTodolistRequest
		if err := json.Unmarshal(trimmed, &inputs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array format: " + err.Error()})
			return
		}
		h.createBulkResponse(c, userID, inputs)
		return
	}

	var wrapped struct {
		Items []CreateTodolistRequest `json:"items"`
		Title string                  `json:"title"`
	}
	if err := json.Unmarshal(trimmed, &wrapped); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON object format: " + err.Error()})
		return
	}

	if len(wrapped.Items) > 0 {
		h.createBulkResponse(c, userID, wrapped.Items)
		return
	}

	created, err := h.service.CreateOne(c.Request.Context(), userID, wrapped.Title)
	if err != nil {
		respondDomainError(c, err, "", "Field 'title' is required", "Failed to create todolist")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Todolist created successfully",
		"data":    ToTodolistResponse(created),
	})
}

func (h *TodolistHandler) createBulkResponse(c *gin.Context, userID uint, inputs []CreateTodolistRequest) {
	titles := make([]string, 0, len(inputs))
	for _, item := range inputs {
		titles = append(titles, item.Title)
	}

	created, err := h.service.CreateBulk(c.Request.Context(), userID, titles)
	if err != nil {
		respondDomainError(c, err, "", "Title cannot be empty for any item", "Failed to create todolists")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Todolists created successfully",
		"count":   len(created),
		"data":    ToTodolistResponseList(created),
	})
}

// PatchOneOrBulk handles PATCH /todolist (single or bulk).
func (h *TodolistHandler) PatchOneOrBulk(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	trimmed, ok := readTrimmedBody(c, "Request body cannot be empty")
	if !ok {
		return
	}

	var itemsToUpdate []UpdateTodolistRequest

	if trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &itemsToUpdate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array format: " + err.Error()})
			return
		}
	} else {
		var wrapped struct {
			Items []UpdateTodolistRequest `json:"items"`
			ID    uint                    `json:"id"`
			Title string                  `json:"title"`
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
			itemsToUpdate = []UpdateTodolistRequest{{
				ID:    wrapped.ID,
				Title: wrapped.Title,
			}}
		}
	}

	if len(itemsToUpdate) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No items provided for update"})
		return
	}

	updates := make([]db.Todolist, len(itemsToUpdate))
	for i, item := range itemsToUpdate {
		updates[i] = db.Todolist{
			ID:    item.ID,
			Title: item.Title,
		}
	}

	updated, err := h.service.UpdateBulk(c.Request.Context(), userID, updates)
	if err != nil {
		respondDomainError(c, err, "One or more todolists not found or unauthorized", "Valid 'id' and 'title' are required for each item", "Failed to update todolists")
		return
	}

	if len(updated) == 1 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Todolist updated successfully",
			"data":    ToTodolistResponse(&updated[0]),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Todolists updated successfully",
		"count":   len(updated),
		"data":    ToTodolistResponseList(updated),
	})
}

// PatchByID handles PATCH /todolist/:id.
func (h *TodolistHandler) PatchByID(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id, ok := parseIDParam(c, "todolist")
	if !ok {
		return
	}

	var req UpdateTodolistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	updated, err := h.service.UpdateByID(c.Request.Context(), id, userID, req.Title)
	if err != nil {
		respondDomainError(c, err, "Todolist not found", "Field 'title' cannot be empty", "Failed to update todolist")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Todolist updated successfully",
		"data":    ToTodolistResponse(updated),
	})
}

// DeleteOneOrBulk handles DELETE /todolist (single or bulk via body).
func (h *TodolistHandler) DeleteOneOrBulk(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	trimmed, ok := readTrimmedBody(c, "Request body is required for deletion")
	if !ok {
		return
	}

	var idsToDelete []uint

	if trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &idsToDelete); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array format: " + err.Error()})
			return
		}
	} else {
		var req DeleteBulkRequest
		if err := json.Unmarshal(trimmed, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}

		if len(req.IDs) > 0 {
			idsToDelete = req.IDs
		} else if req.ID > 0 {
			idsToDelete = []uint{req.ID}
		}
	}

	if len(idsToDelete) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No ID or IDs provided for deletion"})
		return
	}

	deletedCount, err := h.service.DeleteBulk(c.Request.Context(), idsToDelete, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todolist(s)"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Todolist(s) deleted successfully",
		"deletedCount": deletedCount,
	})
}

// DeleteByID handles DELETE /todolist/:id.
func (h *TodolistHandler) DeleteByID(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id, ok := parseIDParam(c, "todolist")
	if !ok {
		return
	}

	rowsAffected, err := h.service.DeleteByID(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todolist"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todolist not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todolist deleted successfully"})
}
