package entities

import "github.com/google/uuid"

type ClusterNode struct {
	Id           int       `json:"id"      db:"id"`
	ClusterId    uuid.UUID `json:"cluster_id"      db:"cluster_id"`
	NodeId       uuid.UUID `json:"node_id" db:"node_id"`
	InstanceName string    `json:"instance_name" db:"instance_name"`
}
