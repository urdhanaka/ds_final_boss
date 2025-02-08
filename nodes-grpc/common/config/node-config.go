package config

import "nodes-grpc/utils"

const (
	NodeServiceGRPCPort = "7000"
)

type NodeIdentification struct {
	Hostname  string `json:"hostname"`
	IpAddress string `json:"ip_address"`
	GrpcPort  string `json:"grpc_port"`
}

func ThisNodeIdentification() NodeIdentification {
    thisNodeIP, err := utils.GetNodeIP()
    if err != nil {
        thisNodeIP = "0.0.0.0"
    }

	thisNode := NodeIdentification{
		Hostname:  utils.GetHostname(),
		IpAddress: thisNodeIP,
        GrpcPort: NodeServiceGRPCPort,
	}

	return thisNode
}
