package model

import (
	"github.com/oklog/ulid/v2"
)

type User struct {
	ID      ulid.ULID `json:"id"`
	Name    string    `json:"name"`
	IsAdmin bool      `json:"is_admin"`
	Group   ulid.ULID `json:"group_id"`
}
