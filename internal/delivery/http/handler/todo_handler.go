package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-gin-todo/internal/delivery/http/dto"
	"go-gin-todo/internal/delivery/http/middleware"
	"go-gin-todo/internal/domain"
	"go-gin-todo/internal/usecase"
)

type TodoHandler struct {
	usecase usecase.TodoUsecase
}

func NewTodoHandler(u usecase.TodoUsecase) *TodoHandler {
	return &TodoHandler{usecase: u}
}

// GetBulk handles GET /todo (bulk) with pagination
func (h *TodoHandler) GetBulk(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	search := c.Query("search")
	checkParam := c.Query("check")
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

	var checkFilter *bool
	if checkParam != "" {
		if b, err := strconv.ParseBool(checkParam); err == nil {
			checkFilter = &b
		}
	}

	todos, pagination, err := h.usecase.GetBulk(c.Request.Context(), userID, search, checkFilter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch todos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       todos,
		"pagination": pagination,
	})
}

// GetOne handles GET /todo/:id
func (h *TodoHandler) GetOne(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}

	todo, err := h.usecase.GetByID(c.Request.Context(), uint(id), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch todo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": todo})
}

// Create handles POST /todo (one, bulk)
func (h *TodoHandler) Create(c *gin.Context) {
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
		var inputs []dto.CreateTodoRequest
		if err := json.Unmarshal(trimmed, &inputs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array format: " + err.Error()})
			return
		}

		var domainTodos []domain.Todo
		for _, item := range inputs {
			isCheck := false
			if item.Check != nil {
				isCheck = *item.Check
			}
			domainTodos = append(domainTodos, domain.Todo{
				Message: item.Message,
				Check:   isCheck,
			})
		}

		created, err := h.usecase.CreateBulk(c.Request.Context(), userID, domainTodos)
		if err != nil {
			if errors.Is(err, domain.ErrBadRequest) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Message cannot be empty for any item"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todos"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Todos created successfully",
			"count":   len(created),
			"data":    created,
		})
		return
	}

	// Object (single or wrapped bulk)
	var wrapped struct {
		Items   []dto.CreateTodoRequest `json:"items"`
		Message string                  `json:"message"`
		Check   *bool                   `json:"check"`
	}
	if err := json.Unmarshal(trimmed, &wrapped); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON object format: " + err.Error()})
		return
	}

	if len(wrapped.Items) > 0 {
		var domainTodos []domain.Todo
		for _, item := range wrapped.Items {
			isCheck := false
			if item.Check != nil {
				isCheck = *item.Check
			}
			domainTodos = append(domainTodos, domain.Todo{
				Message: item.Message,
				Check:   isCheck,
			})
		}

		created, err := h.usecase.CreateBulk(c.Request.Context(), userID, domainTodos)
		if err != nil {
			if errors.Is(err, domain.ErrBadRequest) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Message cannot be empty for any item"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todos"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Todos created successfully",
			"count":   len(created),
			"data":    created,
		})
		return
	}

	// Single Item
	isCheck := false
	if wrapped.Check != nil {
		isCheck = *wrapped.Check
	}

	created, err := h.usecase.CreateOne(c.Request.Context(), userID, wrapped.Message, isCheck)
	if err != nil {
		if errors.Is(err, domain.ErrBadRequest) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Field 'message' is required"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todo"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Todo created successfully",
		"data":    created,
	})
}

// PatchOneOrBulk handles PATCH /todo (update by id, one, bulk)
func (h *TodoHandler) PatchOneOrBulk(c *gin.Context) {
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

	var itemsToUpdate []dto.UpdateTodoRequest

	if trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &itemsToUpdate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array format: " + err.Error()})
			return
		}
	} else {
		var wrapped struct {
			Items   []dto.UpdateTodoRequest `json:"items"`
			ID      uint                    `json:"id"`
			Message *string                 `json:"message"`
			Check   *bool                   `json:"check"`
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
			itemsToUpdate = []dto.UpdateTodoRequest{{
				ID:      wrapped.ID,
				Message: wrapped.Message,
				Check:   wrapped.Check,
			}}
		}
	}

	if len(itemsToUpdate) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No items provided for update"})
		return
	}

	var domainUpdates []domain.TodoUpdateItem
	for _, item := range itemsToUpdate {
		domainUpdates = append(domainUpdates, domain.TodoUpdateItem{
			ID:      item.ID,
			Message: item.Message,
			Check:   item.Check,
		})
	}

	updated, err := h.usecase.UpdateBulk(c.Request.Context(), userID, domainUpdates)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "One or more todos not found or unauthorized"})
			return
		}
		if errors.Is(err, domain.ErrBadRequest) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Valid 'id' is required for each item"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update todos"})
		return
	}

	if len(updated) == 1 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Todo updated successfully",
			"data":    updated[0],
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Todos updated successfully",
			"count":   len(updated),
			"data":    updated,
		})
	}
}

// PatchByID handles PATCH /todo/:id
func (h *TodoHandler) PatchByID(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}

	var req struct {
		Message *string `json:"message"`
		Check   *bool   `json:"check"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.usecase.UpdateByID(c.Request.Context(), uint(id), userID, req.Message, req.Check)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
			return
		}
		if errors.Is(err, domain.ErrBadRequest) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "At least one of 'message' or 'check' must be provided"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update todo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Todo updated successfully",
		"data":    updated,
	})
}

// DeleteOneOrBulk handles DELETE /todo (one, bulk)
// Deletion IDs must strictly be provided in the request body.
func (h *TodoHandler) DeleteOneOrBulk(c *gin.Context) {
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
		var payload dto.DeleteTodoRequest
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Todo(s) deleted successfully",
		"deletedCount": deletedCount,
		"requestedIDs": idsToDelete,
	})
}

// DeleteByID handles DELETE /todo/:id
func (h *TodoHandler) DeleteByID(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}

	deletedCount, err := h.usecase.DeleteByID(c.Request.Context(), uint(id), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todo"})
		return
	}

	if deletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Todo deleted successfully",
		"id":      id,
	})
}
