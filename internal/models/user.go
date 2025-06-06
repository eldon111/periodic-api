package models

// User represents the data model for user objects
type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash []byte `json:"passwordHash"`
}
