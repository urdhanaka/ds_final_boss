package entity

import "github.com/google/uuid"

type Node struct {
	ID        uuid.UUID
	Hostname  string
	IpAddress string
}
