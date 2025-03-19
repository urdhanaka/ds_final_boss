package handlers

import (
	"errors"
	"log/slog"
	"nodes-grpc-frontend/utils"

	proto_model "nodes-grpc-frontend/common/model/proto-model"
	web_model "nodes-grpc-frontend/common/model/web-model"
	usecases "nodes-grpc-frontend/use-cases"

	"github.com/gofiber/fiber/v3"
)

type NodeHandlerImpl interface {
	GetLocalStorage(c fiber.Ctx) *proto_model.NodeList
	Register(c fiber.Ctx) error
	GetNodes(c fiber.Ctx) error
	GetNodesStatus(c fiber.Ctx) error
	CreateCluster(c fiber.Ctx) error
	TestHtmx(c fiber.Ctx) error
}

type nodeHandler struct {
	nodeUsecase usecases.NodeUsecase
}

func NewNodeHandler(nodeUsecase usecases.NodeUsecase) NodeHandlerImpl {
	return &nodeHandler{
		nodeUsecase: nodeUsecase,
	}
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
			"error", err.Error(),
		)

		return err
	}

	err = h.nodeUsecase.RegisterNode(c.Context(), reqBody)
	if err != nil {
		slog.Error("Could not register node",
			"error", err.Error(),
		)

		return err
	}

	return c.SendString("Register node successful")
}

func (h *nodeHandler) GetNodes(c fiber.Ctx) error {
	nodes := h.nodeUsecase.GetAllNodes(c.Context())

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

func (h *nodeHandler) GetNodesStatus(c fiber.Ctx) error {
	nodesStatus, err := h.nodeUsecase.CheckNodesHealth(c.Context())
	if err != nil {
		slog.Error("Could not check nodes status",
			"error", err.Error(),
		)

		return c.SendString("could not check nodes status")
	}

	return c.JSON(nodesStatus)
}

func (h *nodeHandler) CreateCluster(c fiber.Ctx) error {
	c.Accepts("application/json")

	createClusterRequest := new(web_model.CreateClusterRequest)
	serverTokenProto := new(proto_model.ServerToken)

	err := c.Bind().Body(createClusterRequest)
	if err != nil && errors.Is(err, fiber.ErrUnprocessableEntity) {
		createClusterRequest.Token = utils.RandomString(16)
	}

	serverTokenProto.Token = createClusterRequest.Token

	err = h.nodeUsecase.CreateCluster(c.Context(), serverTokenProto)
	if err != nil {
		return err
	}

	return nil
}

func (h *nodeHandler) TestHtmx(c fiber.Ctx) error {
	test := new(web_model.NodeRequirement)

	err := c.Bind().JSON(test)
	if err != nil {
		return err
	}

	return nil
}
