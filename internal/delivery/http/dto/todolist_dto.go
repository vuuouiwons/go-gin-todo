package dto

type CreateTodolistRequest struct {
	Title string `json:"title" binding:"required"`
}

type UpdateTodolistRequest struct {
	ID    uint   `json:"id"`
	Title string `json:"title" binding:"required"`
}
