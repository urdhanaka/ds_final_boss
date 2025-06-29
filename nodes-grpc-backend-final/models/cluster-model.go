package models

import (
	"time"

	"github.com/google/uuid"
)

type AddCluster struct {
	UserId      uuid.UUID `json:"user_id"`
	GroupId     int       `json:"group_id"`
	ClusterId   uuid.UUID `json:"cluster_id"`
	ClusterName string    `json:"cluster_name" binding:"required"`
	Vcpu        int       `json:"vcpu"`
	Memory      int       `json:"memory"`
	Storage     int       `json:"storage"`
	NodeSize    int       `json:"node_size"`
	Nodes       []string  `json:"nodes" binding:"required"`
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

type CreateClusterResponse struct {
	DashboardToken  string `json:"dashboard_token"`
	MasterIpAddress string `json:"master_ip_address"`
}
