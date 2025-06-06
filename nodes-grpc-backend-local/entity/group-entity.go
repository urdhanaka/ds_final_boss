package entity

import "github.com/google/uuid"

type Group struct {
	GroupID        uuid.UUID
	Name           string
	Cpu            int
	Ram            int
	Storage        int
	CurrentCluster int
	MaxCluster     int
}
