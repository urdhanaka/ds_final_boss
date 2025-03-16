package config

import (
	"log/slog"
	"nodes-grpc/common/model"
	"nodes-grpc/utils"
)

const (
	NodeServiceGRPCPort = "7000" // port for running the grpc service
)

type NodeIdentification struct {
	Hostname  string `json:"hostname"`
	IpAddress string `json:"ip_address"`
	GrpcPort  string `json:"grpc_port"`
}

func ThisNodeIdentification() (NodeIdentification, error) {
	thisNodeIP, err := utils.GetNodeIP()
	if err != nil {
		slog.Error("Could not get node IP address",
			"error", err.Error(),
		)

		return NodeIdentification{}, err
	}
	thisNodeHostname, err := utils.GetHostname()
	if err != nil {
		slog.Error("Could not get node IP hostname",
			"error", err.Error(),
		)

		return NodeIdentification{}, err
	}

	thisNode := NodeIdentification{
		Hostname:  thisNodeHostname,
		IpAddress: thisNodeIP,
		GrpcPort:  NodeServiceGRPCPort,
	}

	return thisNode, nil
}

func NewNodeList() *model.NodeList {
	nodeList := new(model.NodeList)
	nodeList.Nodes = make([]*model.NodeStatus, 0)

	return nodeList
}
