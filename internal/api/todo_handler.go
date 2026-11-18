package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-gin-todo/internal/app"
)

// TodoHandler handles HTTP requests for Todo resources.
type TodoHandler struct {
	commands *app.TodoCommands
	queries  *app.TodoQueries
}

// NewTodoHandler constructs a TodoHandler backed by TodoCommands and TodoQueries.
func NewTodoHandler(commands *app.TodoCommands, queries *app.TodoQueries) *TodoHandler {
	return &TodoHandler{
		commands: commands,
		queries:  queries,
	}
}

// GetBulk handles GET /todo with pagination.
func (h *TodoHandler) GetBulk(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	search := c.Query("search")
	checkParam := c.Query("check")
	page, limit := parsePagination(c)

	var checkFilter *bool
	if checkParam != "" {
		if b, err := strconv.ParseBool(checkParam); err == nil {
			checkFilter = &b
		}
	}

	todos, pagination, err := h.queries.GetBulk(c.Request.Context(), userID, search, checkFilter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch todos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       ToTodoResponseList(todos),
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

	todo, err := h.queries.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		respondDomainError(c, err, "Todo not found", "", "Failed to fetch todo")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ToTodoResponse(todo)})
}

// Create handles POST /todo (one or bulk).
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

	created, err := h.commands.CreateOne(c.Request.Context(), userID, wrapped.Message, isCheck)
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
	domainTodos := make([]app.Todo, 0, len(inputs))
	for _, item := range inputs {
		isCheck := false
		if item.Check != nil {
			isCheck = *item.Check
		}
		domainTodos = append(domainTodos, app.Todo{
			Message: item.Message,
			Check:   isCheck,
		})
	}

	created, err := h.commands.CreateBulk(c.Request.Context(), userID, domainTodos)
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

// PatchOneOrBulk handles PATCH /todo (one or bulk).
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

	domainUpdates := make([]app.TodoUpdateItem, 0, len(itemsToUpdate))
	for _, item := range itemsToUpdate {
		domainUpdates = append(domainUpdates, app.TodoUpdateItem{
			ID:      item.ID,
			Message: item.Message,
			Check:   item.Check,
		})
	}

	updated, err := h.commands.UpdateBulk(c.Request.Context(), userID, domainUpdates)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.commands.UpdateByID(c.Request.Context(), id, userID, req.Message, req.Check)
	if err != nil {
		respondDomainError(c, err, "Todo not found", "At least one of 'message' or 'check' must be provided", "Failed to update todo")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Todo updated successfully",
		"data":    ToTodoResponse(updated),
	})
}

// DeleteOneOrBulk handles DELETE /todo strictly from request body IDs.
func (h *TodoHandler) DeleteOneOrBulk(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idsToDelete, ok := parseDeleteIDs(c)
	if !ok {
		return
	}

	deletedCount, err := h.commands.DeleteBulk(c.Request.Context(), idsToDelete, userID)
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

	deletedCount, err := h.commands.DeleteByID(c.Request.Context(), id, userID)
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
