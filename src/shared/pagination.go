package shared

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage     = 1
	DefaultPageStr  = "1"
	DefaultLimit    = 10
	DefaultLimitStr = "10"
)

// Pagination holds computed pagination metadata for list responses.
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}

// ParsePagination extracts and normalizes page and limit from Gin query parameters.
func ParsePagination(c *gin.Context) (int, int) {
	page, err := strconv.Atoi(c.DefaultQuery("page", DefaultPageStr))
	if err != nil || page <= 0 {
		page = DefaultPage
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", DefaultLimitStr))
	if err != nil || limit <= 0 {
		limit = DefaultLimit
	}

	return page, limit
}

// CalculatePagination builds standard pagination metadata.
func CalculatePagination(page, limit int, totalItems int64) Pagination {
	if limit <= 0 {
		limit = DefaultLimit
	}
	if page <= 0 {
		page = DefaultPage
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(limit)))
	}

	return Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1 && totalPages > 0,
	}
}
