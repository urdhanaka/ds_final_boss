package model

import (
	"github.com/oklog/ulid/v2"
)

type Group struct {
	ID         ulid.ULID `json:"id"`
	Name       string    `json:"name"`
	MaxCpu     int       `json:"max_cpu"`
	MaxRam     int       `json:"max_ram"`
	MaxStorage int       `json:"max_storage"`
}
