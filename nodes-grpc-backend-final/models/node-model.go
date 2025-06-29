package models

import "github.com/google/uuid"

type GetGroupNodes struct {
	NodeId    uuid.UUID `json:"node_id"`
	Hostname  string    `json:"hostname"`
	IpAddress string    `json:"ip_address"`
	VCpu      int       `json:"vcpu"`
	Memory    int       `json:"memory"`
	Storage   int       `json:"storage"`
	GroupId   int       `json:"group_id"`
}

type AddNode struct {
	Hostname  string `json:"hostname"`
	IpAddress string `json:"ip_address"`
	LabName   string `json:"lab_name"`
	VCpu      int    `json:"vcpu"`
	Storage   int    `json:"storage"`
	Memory    int    `json:"memory"`
}
