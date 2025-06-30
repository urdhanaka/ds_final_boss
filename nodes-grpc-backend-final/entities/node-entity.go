package entities

import "github.com/google/uuid"

type Node struct {
	NodeID    uuid.UUID `json:"node_id,omitempty" db:"node_id"`       // node/worker ID
	Hostname  string    `json:"hostname,omitempty" db:"hostname"`     // hostname
	IpAddress string    `json:"ip_address,omitempty" db:"ip_address"` // ip address
	Memory    int       `json:"memory" db:"memory"`
	VCpu      int       `json:"vcpu" db:"vcpu"`
	Storage   int       `json:"storage" db:"storage"`
	GroupId   int       `json:"group_id,omitempty" db:"group_id"` // group ID
}
