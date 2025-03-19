package utils

import (
	"errors"
	"math/rand"
	proto_model "nodes-grpc-frontend/common/model/proto-model"
	"slices"
	"time"
)

func RandomString(length int) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, length)
	for i := range b {
		b[i] = letters[random.Intn(len(letters))]
	}

	return string(b)
}

// Get random node from list of nodes
//
// The selected node will also popped from the list
func GetRandomNode(listOfNodes *proto_model.NodeList) (*proto_model.NodeStatus, error) {
	if listOfNodes == nil || listOfNodes.Nodes == nil {
		return new(proto_model.NodeStatus), errors.New("list of nodes is nil")
	}

	numberOfNodes := len(listOfNodes.Nodes)
	selectedInteger := rand.Intn(numberOfNodes)
	selectedNode := listOfNodes.Nodes[selectedInteger]

	return selectedNode, nil
}

func DeleteFromPool(listOfNodes *proto_model.NodeList, selectedNode *proto_model.NodeStatus) error {
	copyOfList := listOfNodes.Nodes

	for index, node := range copyOfList {
		if node.IpAddress == selectedNode.IpAddress {
			copyOfList = slices.Delete(copyOfList, index, index+1)

			break
		}
	}

	listOfNodes.Nodes = copyOfList

	return nil
}
