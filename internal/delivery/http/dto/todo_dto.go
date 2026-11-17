package dto

type CreateTodoRequest struct {
	Message string `json:"message" binding:"required"`
	Check   *bool  `json:"check"`
}

type UpdateTodoRequest struct {
	ID      uint    `json:"id"`
	Message *string `json:"message"`
	Check   *bool   `json:"check"`
}

type DeleteTodoRequest struct {
	ID  uint   `json:"id"`
	IDs []uint `json:"ids"`
}
