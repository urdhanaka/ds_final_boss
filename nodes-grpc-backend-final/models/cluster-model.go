package models

import "github.com/google/uuid"

type AddCluster struct {
	UserId      uuid.UUID `json:"user_id"      binding:"required"`
	GroupId     int       `json:"group_id"     binding:"required"`
	ClusterName string    `json:"cluster_name" binding:"required"`
}
