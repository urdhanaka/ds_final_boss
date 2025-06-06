package entity

import (
	"time"

	"github.com/google/uuid"
)

type Cluster struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Group     uuid.UUID
	CreatedAt time.Time
}
