package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"nodes-grpc-frontend/common/model/web"
	"nodes-grpc-frontend/services/node_service"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type NodeHandlerImpl interface {
	Register(c fiber.Ctx) error
	GetNodes(c fiber.Ctx) []web.Node
	CreateCluster(c fiber.Ctx) error
	GetDashboard(c fiber.Ctx) (string, error)
	TempGetDashboard(c fiber.Ctx) error
	// GetNodesStatus(c fiber.Ctx) error
}

type NodeHandler struct {
	nodeUsecase node_service.NodeUsecase
}

func NewNodeHandler(nodeService node_service.NodeUsecase) NodeHandlerImpl {
	return &NodeHandler{
		nodeUsecase: nodeService,
	}
}

func (h *NodeHandler) Register(c fiber.Ctx) error {
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

func (h *NodeHandler) GetNodes(c fiber.Ctx) []web.Node {
	nodes, err := h.nodeUsecase.GetAllNodes(c.Context())
	if err != nil {
		slog.Error("GetNodes(): error getting nodes",
			"error", err.Error(),
		)
	}

	return nodes
}

func (h *NodeHandler) CreateCluster(c fiber.Ctx) error {
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

	return c.SendString("test")
}

func (h *NodeHandler) GetDashboard(c fiber.Ctx) (string, error) {
	token := c.Params("token")
	masterNode, err := h.nodeUsecase.AccessCluster(c.Context(), token)
	if err != nil {
		slog.Error("Dashboard(): Could not get request body",
			"error", err.Error(),
		)

		return "", err
	}

	// cluster doesn't exists
	if masterNode.NodeIP == "" {
		return "", errors.New("Cluster doesn't exists")
	}

	return masterNode.NodeIP, nil
}

func (h *NodeHandler) TempGetDashboard(c fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).SendString("token must be provided")
	}

	masterNode, err := h.nodeUsecase.AccessCluster(c.Context(), token)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Cluster not found")
	}

	prefix := "/dashboard/" + token
	handler := reverseProxy(masterNode.NodeIP, prefix)

	return handler(c)
}

func reverseProxy(targetHost, prefix string) fiber.Handler {
	targetUrl, err := url.Parse("https://" + targetHost)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	originalDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		originalDirector(r)
		r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
		r.Host = targetUrl.Host
	}

	return func(c fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(proxy)
		return nil
	}
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
