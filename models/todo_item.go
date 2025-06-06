package models

// TodoItem represents a to-do item with a text description and checked status
type TodoItem struct {
	ID      int64  `json:"id"`
	Text    string `json:"text"`
	Checked bool   `json:"checked"`
}