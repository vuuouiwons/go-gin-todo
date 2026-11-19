package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-gin-todo/src/db"
	"go-gin-todo/src/services"
	"go-gin-todo/src/utils"
)

// TodoHandler handles HTTP requests for Todo resources.
type TodoHandler struct {
	service *services.TodoService
}

// NewTodoHandler constructs a new TodoHandler.
func NewTodoHandler(service *services.TodoService) *TodoHandler {
	return &TodoHandler{service: service}
}

// GetBulk handles GET /todo.
func (h *TodoHandler) GetBulk(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	search := c.Query("search")
	checkParam := c.Query("check")
	page, limit := utils.ParsePagination(c)

	var checkFilter *bool
	if checkParam != "" {
		if b, err := strconv.ParseBool(checkParam); err == nil {
			checkFilter = &b
		}
	}

	items, pagination, err := h.service.GetBulk(c.Request.Context(), userID, search, checkFilter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch todos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       ToTodoResponseList(items),
		"pagination": ToPaginationResponse(pagination),
	})
}

// GetOne handles GET /todo/:id.
func (h *TodoHandler) GetOne(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id, ok := parseIDParam(c, "todo")
	if !ok {
		return
	}

	item, err := h.service.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		respondDomainError(c, err, "Todo not found", "", "Failed to fetch todo")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ToTodoResponse(item)})
}

// Create handles POST /todo (single or bulk).
func (h *TodoHandler) Create(c *gin.Context) {
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
		var inputs []CreateTodoRequest
		if err := json.Unmarshal(trimmed, &inputs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array format: " + err.Error()})
			return
		}
		h.createBulkResponse(c, userID, inputs)
		return
	}

	var wrapped struct {
		Items   []CreateTodoRequest `json:"items"`
		Message string              `json:"message"`
		Check   *bool               `json:"check"`
	}
	if err := json.Unmarshal(trimmed, &wrapped); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON object format: " + err.Error()})
		return
	}

	if len(wrapped.Items) > 0 {
		h.createBulkResponse(c, userID, wrapped.Items)
		return
	}

	isCheck := false
	if wrapped.Check != nil {
		isCheck = *wrapped.Check
	}

	created, err := h.service.CreateOne(c.Request.Context(), userID, wrapped.Message, isCheck)
	if err != nil {
		respondDomainError(c, err, "", "Field 'message' is required", "Failed to create todo")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Todo created successfully",
		"data":    ToTodoResponse(created),
	})
}

func (h *TodoHandler) createBulkResponse(c *gin.Context, userID uint, inputs []CreateTodoRequest) {
	domainTodos := make([]db.Todo, 0, len(inputs))
	for _, item := range inputs {
		isCheck := false
		if item.Check != nil {
			isCheck = *item.Check
		}
		domainTodos = append(domainTodos, db.Todo{
			Message: item.Message,
			Check:   isCheck,
		})
	}

	created, err := h.service.CreateBulk(c.Request.Context(), userID, domainTodos)
	if err != nil {
		respondDomainError(c, err, "", "Message cannot be empty for any item", "Failed to create todos")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Todos created successfully",
		"count":   len(created),
		"data":    ToTodoResponseList(created),
	})
}

// PatchOneOrBulk handles PATCH /todo (single or bulk).
func (h *TodoHandler) PatchOneOrBulk(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	trimmed, ok := readTrimmedBody(c, "Request body cannot be empty")
	if !ok {
		return
	}

	var itemsToUpdate []UpdateTodoRequest

	if trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &itemsToUpdate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array format: " + err.Error()})
			return
		}
	} else {
		var wrapped struct {
			Items   []UpdateTodoRequest `json:"items"`
			ID      uint                `json:"id"`
			Message *string             `json:"message"`
			Check   *bool               `json:"check"`
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
			itemsToUpdate = []UpdateTodoRequest{{
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

	updates := make([]db.TodoUpdateItem, len(itemsToUpdate))
	for i, item := range itemsToUpdate {
		updates[i] = db.TodoUpdateItem{
			ID:      item.ID,
			Message: item.Message,
			Check:   item.Check,
		}
	}

	updated, err := h.service.UpdateBulk(c.Request.Context(), userID, updates)
	if err != nil {
		respondDomainError(c, err, "One or more todos not found or unauthorized", "Valid 'id' is required for each item", "Failed to update todos")
		return
	}

	if len(updated) == 1 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Todo updated successfully",
			"data":    ToTodoResponse(&updated[0]),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Todos updated successfully",
		"count":   len(updated),
		"data":    ToTodoResponseList(updated),
	})
}

// PatchByID handles PATCH /todo/:id.
func (h *TodoHandler) PatchByID(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id, ok := parseIDParam(c, "todo")
	if !ok {
		return
	}

	var req struct {
		Message *string `json:"message"`
		Check   *bool   `json:"check"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	updated, err := h.service.UpdateByID(c.Request.Context(), id, userID, req.Message, req.Check)
	if err != nil {
		respondDomainError(c, err, "Todo not found", "Message cannot be empty", "Failed to update todo")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Todo updated successfully",
		"data":    ToTodoResponse(updated),
	})
}

// DeleteOneOrBulk handles DELETE /todo (single or bulk via body).
func (h *TodoHandler) DeleteOneOrBulk(c *gin.Context) {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todo(s)"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Todo(s) deleted successfully",
		"deletedCount": deletedCount,
	})
}

// DeleteByID handles DELETE /todo/:id.
func (h *TodoHandler) DeleteByID(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id, ok := parseIDParam(c, "todo")
	if !ok {
		return
	}

	rowsAffected, err := h.service.DeleteByID(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todo"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}
