package utils

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-gin-todo/src/db"
)

const (
	DefaultPage     = 1
	DefaultPageStr  = "1"
	DefaultLimit    = 10
	DefaultLimitStr = "10"
)

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
func CalculatePagination(page, limit int, totalItems int64) db.Pagination {
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

	return db.Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1 && totalPages > 0,
	}
}
