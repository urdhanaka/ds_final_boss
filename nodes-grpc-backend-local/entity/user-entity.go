package entity

import "github.com/google/uuid"

type User struct {
	UserId   uuid.UUID `json:"user_id"  db:"user_id"`  // primary key
	Name     string    `json:"name"     db:"name"`     // user name
	Email    string    `json:"string"   db:"email"`    // user email
	Password string    `json:"password" db:"password"` // user password
	GroupID  uuid.UUID `json:"group_id" db:"group_id"` // group id
	Role     string    `json:"role"     db:"role"`     // is admin or not
}
