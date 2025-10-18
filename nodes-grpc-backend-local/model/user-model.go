package model

import "github.com/google/uuid"

type AddUser struct {
	Name     string    `json:"name"     binding:"required"`
	Email    string    `json:"email"    binding:"required"`
	Password string    `json:"password" binding:"required"`
	GroupId  uuid.UUID `json:"group_id" binding:"required"`
	IsAdmin  bool      `json:"is_admin"`
}

type LoginUser struct {
	Email    string `json:"email"       binding:"required"`
	Password string `json:"password"    binding:"required"`
}
