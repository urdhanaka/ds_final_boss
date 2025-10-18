package entity

import "github.com/google/uuid"

type Node struct {
	ID        uuid.UUID `json:"id"`
	Hostname  string    `json:"hostname"`
	IpAddress string    `json:"ip_address"`
}
