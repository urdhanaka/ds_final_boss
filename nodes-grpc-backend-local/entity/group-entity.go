package entity

type Group struct {
	GroupId    int    `json:"group_id"    db:"group_id"` // primary key
	Name       string `json:"name"        db:"name"`     // name of the group (AJK, NCC, RPL, KCV, etc..)
	Vcpu       int    `json:"vcpu"        db:"vcpu"`     // max vcpu that can be used by this group
	Ram        int    `json:"ram"         db:"ram"`      // max ram size that can be used by this group
	Storage    int    `json:"storage"     db:"storage"`  // max storage size that can be used by this group
	NodeSize   int    `json:"node_size"   db:"node_size"`
	MaxCluster int    `json:"max_cluster" db:"max_cluster"` // max cluster that can be created under this group
}
