package entity

import "github.com/google/uuid"

type User struct {
	UserID  uuid.UUID `json:"user_id,omitempty"`
	Name    string    `json:"name,omitempty"`
	GroupID uuid.UUID `json:"group_id,omitempty"`
	IsAdmin bool      `json:"is_admin,omitempty"`
}
