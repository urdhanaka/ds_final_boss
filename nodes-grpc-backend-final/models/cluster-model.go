package models

import (
	"time"

	"github.com/google/uuid"
)

type AddCluster struct {
	UserId      uuid.UUID `json:"user_id"      binding:"required"`
	GroupId     int       `json:"group_id"     binding:"required"`
	ClusterName string    `json:"cluster_name" binding:"required"`
}

type GetUserClusters struct {
	ClusterId     uuid.UUID `json:"cluster_id"`
	ClusterName   string    `json:"cluster_name"`
	ClusterStatus string    `json:"cluster_status"`
}

type GetClusterDetails struct {
	ClusterId     uuid.UUID `json:"cluster_id"`
	ClusterName   string    `json:"cluster_name"`
	UserId        uuid.UUID `json:"user_id"`
	ClusterStatus string    `json:"cluster_status"`
	IpAddress     string    `json:"ip_address"`
	AccessToken   string    `json:"access_token"`
	CreatedAt     time.Time `json:"created_at"`
}
