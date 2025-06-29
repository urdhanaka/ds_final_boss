package entities

type Group struct {
	GroupId        int    `json:"group_id"    db:"group_id"`            // primary key
	Name           string `json:"name"        db:"name"`                // name of the group (AJK, NCC, RPL, KCV, etc..)
	Vcpu           int    `json:"vcpu"        db:"vcpu"`                // max vcpu that can be used by this group
	Memory         int    `json:"memory"      db:"memory"`              // max ram size that can be used by this group
	Storage        int    `json:"storage"     db:"storage"`             // max storage size that can be used by this group
	NodeSize       int    `json:"node_size"   db:"node_size"`           // node amount per cluster
	CurrentCluster int    `json:"current_cluster" db:"current_cluster"` // current cluster that exists
	MaxCluster     int    `json:"max_cluster" db:"max_cluster"`         // max cluster that can be created under this group
}
