package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	proto_model "nodes-grpc-frontend/common/model/proto-model"
	web_model "nodes-grpc-frontend/common/model/web-model"
	usecases "nodes-grpc-frontend/use-cases"
	"nodes-grpc-frontend/utils"

	"github.com/gofiber/fiber/v3"
)

type NodeHandlerImpl interface {
	GetLocalStorage(c fiber.Ctx) *proto_model.NodeList
	Register(c fiber.Ctx) error
	GetNodes(c fiber.Ctx) error
	CreateCluster(c fiber.Ctx) error
}

type nodeHandler struct {
	nodeUsecase usecases.NodeUsecase
}

func NewNodeHandler(nodeUsecase usecases.NodeUsecase) NodeHandlerImpl {
	return &nodeHandler{
		nodeUsecase: nodeUsecase,
	}
}

func setHeader(c fiber.Ctx) {
    c.Set("", "")
    c.Set("", "")

    return
}

func (h *nodeHandler) GetLocalStorage(c fiber.Ctx) *proto_model.NodeList {
	return h.nodeUsecase.GetAllNodes(c.Context())
}

func (h *nodeHandler) Register(c fiber.Ctx) error {
	// accepts only json request
	c.Accepts("application/json")

	reqBody := new(web_model.RegisterNodeRequest)

	err := c.Bind().Body(reqBody)
	if err != nil {
		slog.Error("Could not get request body",
			"error", err)

		return err
	}

	err = h.nodeUsecase.RegisterNode(c.Context(), reqBody)
	if err != nil {
		slog.Error("Could not register node",
			"error", err)

		return err
	}

	return c.SendString("Register node successful")
}

func (h *nodeHandler) GetNodes(c fiber.Ctx) error {
	nodes := h.nodeUsecase.GetAllNodes(c.Context())

	// response := new([]web_model.GetNodesResponse)
	response := make([]web_model.GetNodesResponse, 0)

	for _, node := range nodes.Nodes {
		thisNode := web_model.GetNodesResponse{
			Hostname:  node.Hostname,
			IpAddress: node.IpAddress,
			GrpcPort:  node.GrpcPort,
		}

		response = append(response, thisNode)
	}

	return c.JSON(response)
}

func (h *nodeHandler) CreateCluster(c fiber.Ctx) error {
	c.Accepts("application/json")

	createClusterRequest := new(web_model.CreateClusterRequest)
	// serverTokenProto := new(proto_model.ServerToken)

	err := c.Bind().Body(createClusterRequest)
	if err != nil && errors.Is(err, fiber.ErrUnprocessableEntity){
        createClusterRequest.Token = utils.RandomString(16)
	}

    fmt.Println(createClusterRequest.Token)

	// serverTokenProto.Token = createClusterRequest.Token

	// err = h.nodeUsecase.CreateCluster(c.Context(), serverTokenProto)
	// if err != nil {
	// 	return err
	// }

	return nil
}
