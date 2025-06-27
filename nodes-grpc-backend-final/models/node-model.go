package models

import "github.com/google/uuid"

type GetGroupNodes struct {
	NodeId    uuid.UUID `json:"node_id"`
	Hostname  string    `json:"hostname"`
	IpAddress string    `json:"ip_address"`
	VCpu      int       `json:"vcpu"`
	Ram       int       `json:"ram"`
	Storage   int       `json:"storage"`
	GroupId   int       `json:"group_id"`
}
