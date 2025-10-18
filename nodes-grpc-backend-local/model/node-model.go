package model

import "github.com/google/uuid"

type AddNode struct {
	Hostname  string `json:"hostname"   binding:"required"`
	IpAddress string `json:"ip_address" binding:"required"`
	Cpu       int    `json:"cpu"        binding:"required"`
	Ram       int    `json:"ram"        binding:"required"`
	Storage   int    `json:"storage"    binding:"required"`
	LabName   string `json:"lab_name"   binding:"required"`
}

type AssignNode struct {
	NodeId  uuid.UUID `json:"node_id"   binding:"required"`
	GroupId uuid.UUID `json:"group_id"  binding:"required"`
}
