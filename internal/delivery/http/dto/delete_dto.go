package dto

// DeleteRequest represents the JSON object body for single or bulk deletion.
// Used across Todolist and Todo deletion endpoints.
type DeleteRequest struct {
	ID  uint   `json:"id"`
	IDs []uint `json:"ids"`
}
