package config

import (
	"log/slog"
	"nodes-grpc/utils"
	"os"
)

type NodeIdentification struct {
	Hostname  string `json:"hostname"`
	IpAddress string `json:"ip_address"`
	GrpcPort  string `json:"grpc_port"`
}

func ThisNodeIdentification() (NodeIdentification, error) {
	thisNodeHostname, err := os.Hostname()
	if err != nil {
		slog.Error("Could not get node hostname",
			"error", err.Error(),
		)

		return NodeIdentification{}, err
	}

	thisNodeIP, err := utils.GetNodeIP()
	if err != nil {
		slog.Error("Could not get node IP address",
			"error", err.Error(),
		)

		return NodeIdentification{}, err
	}

	thisNode := NodeIdentification{
		Hostname:  thisNodeHostname,
		IpAddress: thisNodeIP,
	}

	return thisNode, nil
}
