package entities

import (
	"time"

	"github.com/google/uuid"
)

type Cluster struct {
	ClusterID     uuid.UUID `json:"cluster_id"      db:"cluster_id"`     // cluster id
	ClusterName   string    `json:"cluster_name"    db:"cluster_name"`   // cluster name
	UserID        uuid.UUID `json:"user_id"         db:"user_id"`        // user that created the cluster
	GroupID       int       `json:"group_id"        db:"group_id"`       // the group the cluster belongs to
	ClusterStatus string    `json:"cluster_status"  db:"cluster_status"` // cluster creation status
	IpAddress     *string   `json:"ip_address"      db:"ip_address"`     // ip address of the master node to access the cluster dashboard
	AccessToken   *string   `json:"access_token"    db:"access_token"`   // access token to access the cluster dashboard
	CreatedAt     time.Time `json:"created_at"      db:"created_at"`     // cluster creation timedate
}
