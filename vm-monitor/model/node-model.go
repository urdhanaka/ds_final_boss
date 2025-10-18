package model

type Node struct {
	IP string `db:"id"`
}

type NodeStatus struct {
	Name     string
	Error    error
	CpuTime  uint64
	IsActive bool
}
