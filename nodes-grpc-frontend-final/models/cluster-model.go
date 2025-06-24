package models

import (
	"time"

	"github.com/google/uuid"
)

type Clusters struct {
	ClusterId     uuid.UUID `json:"cluster_id"`
	ClusterName   string    `json:"cluster_name"`
	ClusterStatus string    `json:"cluster_status"`
	IpAddress     string    `json:"ip_address"`
	AccessToken   string    `json:"access_token"`
	CreatedAt     time.Time `json:"created_at"`
}
