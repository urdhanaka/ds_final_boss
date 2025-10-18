package web

type Node struct {
	UUID     string `json:"uuid" db:"uuid"`
	NodeIP   string `json:"node_ip" db:"node_ip"`
	GrpcPort string `json:"grpc_port" db:"grpc_port"`
}

type Dashboard struct {
	UUID   string `json:"uuid" db:"uuid"`
	NodeIP string `json:"node_ip" db:"node_ip"`
}
