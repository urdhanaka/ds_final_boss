package handlers

import (
	"fmt"
	"log/slog"
	"nodes-grpc-frontend/common/model/web"
	"nodes-grpc-frontend/services/db_service"
	"nodes-grpc-frontend/services/node_service"

	"github.com/gofiber/fiber/v3"
)

type NodeHandlerImpl interface {
	Register(c fiber.Ctx) error
	GetNodes(c fiber.Ctx) []db_service.Node
	CreateCluster(c fiber.Ctx) error
	// GetNodesStatus(c fiber.Ctx) error
}

type nodeHandler struct {
	nodeUsecase node_service.NodeUsecase
}

func NewNodeHandler(nodeService node_service.NodeUsecase) NodeHandlerImpl {
	return &nodeHandler{
		nodeUsecase: nodeService,
	}
}

func (h *nodeHandler) Register(c fiber.Ctx) error {
	// accepts only json request
	c.Accepts("application/json")

	fmt.Println(c.GetReqHeaders())

	reqBody := new(web.RegisterNodeRequest)

	err := c.Bind().Body(reqBody)
	if err != nil {
		slog.Error("Register(): Could not get request body",
			"error", err.Error(),
		)

		return err
	}

	err = h.nodeUsecase.RegisterNode(c.Context(), reqBody)
	if err != nil {
		slog.Error("RegisterNode(): Could not register node",
			"error", err.Error(),
		)

		return err
	}

	return c.SendString("Register node successful")
}

func (h *nodeHandler) GetNodes(c fiber.Ctx) []db_service.Node {
	nodes, err := h.nodeUsecase.GetAllNodes(c.Context())
	if err != nil {
		slog.Error("GetNodes(): error getting nodes",
			"error", err.Error(),
		)
	}

	return nodes
}

func (h *nodeHandler) CreateCluster(c fiber.Ctx) error {
	c.Accepts("application/json")

	fmt.Println(c.GetReqHeaders())

	reqBody := new(web.CreateClusterRequest)

	err := c.Bind().Body(reqBody)
	if err != nil {
		slog.Error("CreateCluster(): Could not get request body",
			"error", err.Error(),
		)

		return err
	}

	fmt.Println(reqBody)

	return c.SendString("test")
}

// func (h *nodeHandler) GetNodesStatus(c fiber.Ctx) error {
// 	nodesStatus, err := h.nodeUsecase.CheckNodesHealth(c.Context())
// 	if err != nil {
// 		slog.Error("Could not check nodes status",
// 			"error", err.Error(),
// 		)
//
// 		return c.SendString("could not check nodes status")
// 	}
//
// 	return c.JSON(nodesStatus)
// }

// func (h *nodeHandler) CreateCluster(c fiber.Ctx) error {
// 	c.Accepts("application/json")
//
// 	reqBody := new(web.NodeRequirement)
// 	err := c.Bind().JSON(reqBody)
// 	if err != nil {
// 		return err
// 	}
//
// 	tokenModel := &proto.ServerToken{
// 		Token: utils.RandomString(16),
// 	}
//
// 	err = h.nodeUsecase.CreateCluster(c.Context(), tokenModel)
//
// 	return nil

// createClusterRequest := new(web_model.CreateClusterRequest)
// serverTokenProto := new(proto_model.ServerToken)
//
// err := c.Bind().Body(createClusterRequest)
// if err != nil && errors.Is(err, fiber.ErrUnprocessableEntity) {
// 	createClusterRequest.Token = utils.RandomString(16)
// }
//
// serverTokenProto.Token = createClusterRequest.Token
//
// err = h.nodeUsecase.CreateCluster(c.Context(), serverTokenProto)
// if err != nil {
// 	return err
// }
//
// return nil
// }
