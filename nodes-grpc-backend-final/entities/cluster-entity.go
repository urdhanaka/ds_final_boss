package entities

import (
	"time"

	"github.com/google/uuid"
)

type Cluster struct {
	ClusterID uuid.UUID // cluster id
	Name      string    // cluster name
	UserID    uuid.UUID // user that created the cluster
	GroupID   int       // the group the cluster belongs to
	CreatedAt time.Time // cluster creation timedate
}
