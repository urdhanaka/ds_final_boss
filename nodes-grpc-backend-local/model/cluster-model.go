package model

import "github.com/google/uuid"

type AddCluster struct {
	UserId      uuid.UUID `json:"user_id"      binding:"required"`
	GroupId     int       `json:"group_id"     binding:"required"`
	ClusterName string    `json:"cluster_name" binding:"required"`
	NodeSize    int       `json:"node_size"    binding:"required"`
	Vcpu        int       `json:"vcpu"         binding:"required"`
	Memory      int       `json:"ram"          binding:"required"`
	Storage     int       `json:"storage"      binding:"required"`
}
