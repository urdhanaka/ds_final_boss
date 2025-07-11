package handlers

import (
	"errors"
	"fmt"
	"nodes-grpc-fe/consts"
	"nodes-grpc-fe/models"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

var (
	ErrTokenIsMissing        = errors.New("token not found")
	ErrUnauthorized          = errors.New("user unauthorized")
	ErrServerIsNotResponding = errors.New("server is not responding")
	ErrClusterIdIsNotValid   = errors.New("cluster id is not valid")
)

type ApiClient struct {
	client *resty.Client
}

func NewApiClient(
	client *resty.Client,
) *ApiClient {
	return &ApiClient{
		client,
	}
}

func (api *ApiClient) isServerUp() bool {
	res, err := api.client.R().
		SetResult(&models.ApiResponse{}).
		Get(consts.BACKEND_URL + "/api/health")
	if err != nil {
		return false
	}

	if !res.IsSuccess() {
		return false
	}

	return true
}

func (api *ApiClient) Login(c *gin.Context, email, password string) (*models.ApiResponse, error) {
	if !api.isServerUp() {
		return nil, fmt.Errorf("server is not responding")
	}

	userLogin := &models.LoginUser{
		Email:    email,
		Password: password,
	}

	resp, err := api.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(userLogin).
		SetResult(&models.ApiResponse{}).
		SetError(&models.ApiResponse{}).
		Post(consts.BACKEND_URL + "/api/users/login")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() < 300 {
		respModel := resp.Result().(*models.ApiResponse)

		return respModel, nil
	} else {
		respModel := resp.Error().(*models.ApiResponse)

		return respModel, nil
	}
}

func (api *ApiClient) Me(c *gin.Context) (*models.ApiResponse, error) {
	if !api.isServerUp() {
		return nil, fmt.Errorf("server is not responding")
	}

	token, tokenExists := c.Get("token")
	if !tokenExists {
		return nil, ErrTokenIsMissing
	}

	bearerToken := fmt.Sprintf("Bearer %s", token.(string))
	resp, err := api.client.R().
		SetHeader("Authorization", bearerToken).
		SetResult(&models.ApiResponse{}).
		SetError(&models.ApiResponse{}).
		Get(consts.BACKEND_URL + "/api/users/me")
	if err != nil {
		return &models.ApiResponse{}, err
	}

	if resp.StatusCode() < 300 {
		respModel := resp.Result().(*models.ApiResponse)

		return respModel, nil
	} else {
		respModel := resp.Error().(*models.ApiResponse)

		return respModel, nil
	}
}

func (api *ApiClient) GetGroupNodes(c *gin.Context) (*models.ApiResponse, error) {
	if !api.isServerUp() {
		return nil, fmt.Errorf("server is not responding")
	}

	token, tokenExists := c.Get("token")
	if !tokenExists {
		return nil, ErrTokenIsMissing
	}

	bearerToken := fmt.Sprintf("Bearer %s", token.(string))
	resp, err := api.client.R().
		SetHeader("Authorization", bearerToken).
		SetResult(&models.ApiResponse{}).
		SetError(&models.ApiResponse{}).
		Get(consts.BACKEND_URL + "/api/nodes/group-nodes")
	if err != nil {
		return &models.ApiResponse{}, err
	}

	if resp.StatusCode() < 300 {
		respModel := resp.Result().(*models.ApiResponse)

		return respModel, nil
	} else {
		respModel := resp.Error().(*models.ApiResponse)

		return respModel, nil
	}
}

func (api *ApiClient) GetUserClusters(c *gin.Context) (*models.ApiResponse, error) {
	if !api.isServerUp() {
		return nil, ErrServerIsNotResponding
	}

	token, tokenExists := c.Get("token")
	if !tokenExists {
		return nil, ErrTokenIsMissing
	}

	bearerToken := fmt.Sprintf("Bearer %s", token.(string))
	resp, err := api.client.R().
		SetHeader("Authorization", bearerToken).
		SetResult(&models.ApiResponse{}).
		Get(consts.BACKEND_URL + "/api/clusters")
	if err != nil {
		return &models.ApiResponse{}, err
	}

	if resp.StatusCode() < 300 {
		respModel := resp.Result().(*models.ApiResponse)

		return respModel, nil
	} else {
		respModel := resp.Error().(*models.ApiResponse)

		return respModel, nil
	}
}

func (api *ApiClient) GetClusterDetails(c *gin.Context, clusterId string) (*models.ApiResponse, error) {
	if !api.isServerUp() {
		return nil, ErrServerIsNotResponding
	}

	token, tokenExists := c.Get("token")
	if !tokenExists {
		return nil, ErrTokenIsMissing
	}

	bearerToken := fmt.Sprintf("Bearer %s", token.(string))
	resp, err := api.client.R().
		SetHeader("Authorization", bearerToken).
		SetResult(&models.ApiResponse{}).
		SetError(&models.ApiResponse{}).
		Get(consts.BACKEND_URL + "/api/clusters/" + clusterId)
	if err != nil {
		return &models.ApiResponse{}, err
	}

	if resp.StatusCode() < 300 {
		respModel := resp.Result().(*models.ApiResponse)

		return respModel, nil
	} else {
		respModel := resp.Error().(*models.ApiResponse)

		return respModel, nil
	}
}

func (api *ApiClient) CreateCluster(
	c *gin.Context,
	createClusterModel *models.CreateCluster,
) (*models.ApiResponse, error) {
	if !api.isServerUp() {
		return nil, ErrServerIsNotResponding
	}

	token, tokenExists := c.Get("token")
	if !tokenExists {
		return nil, ErrTokenIsMissing
	}

	bearerToken := fmt.Sprintf("Bearer %s", token.(string))
	resp, err := api.client.R().
		SetHeader("Authorization", bearerToken).
		SetResult(&models.ApiResponse{}).
		SetError(&models.ApiResponse{}).
		SetBody(createClusterModel).
		Post(consts.BACKEND_URL + "/api/clusters")
	if err != nil {
		return &models.ApiResponse{}, err
	}

	if resp.StatusCode() < 300 {
		respModel := resp.Result().(*models.ApiResponse)

		return respModel, nil
	} else {
		respModel := resp.Error().(*models.ApiResponse)

		return respModel, nil
	}
}

func (api *ApiClient) DownloadKubeconfig(
	c *gin.Context,
	clusterId string,
) error {
	if !api.isServerUp() {
		return ErrServerIsNotResponding
	}

	token, tokenExists := c.Get("token")
	if !tokenExists {
		return ErrTokenIsMissing
	}

	bearerToken := fmt.Sprintf("Bearer %s", token.(string))
	resp, err := api.client.R().
		SetHeader("Authorization", bearerToken).
		Get(consts.BACKEND_URL + "/api/clusters/" + clusterId + "/kubeconfig")
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to download kubeconfig: %s", resp.Status())
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s-kubeconfig.yaml", clusterId))
	c.Data(200, "application/octet-stream", resp.Body())

	return nil
}

func (api *ApiClient) DeleteCluster(
	c *gin.Context,
	deleteClusterModel *models.DeleteCluster,
) (*models.ApiResponse, error) {
	if !api.isServerUp() {
		return nil, ErrServerIsNotResponding
	}

	token, tokenExists := c.Get("token")
	if !tokenExists {
		return nil, ErrTokenIsMissing
	}

	bearerToken := fmt.Sprintf("Bearer %s", token.(string))
	resp, err := api.client.R().
		SetHeader("Authorization", bearerToken).
		SetResult(&models.ApiResponse{}).
		SetError(&models.ApiResponse{}).
		Delete(consts.BACKEND_URL + "/api/clusters/" + deleteClusterModel.ClusterId.String())
	if err != nil {
		return &models.ApiResponse{}, err
	}

	if resp.StatusCode() < 300 {
		respModel := resp.Result().(*models.ApiResponse)

		return respModel, nil
	} else {
		respModel := resp.Error().(*models.ApiResponse)

		return respModel, nil
	}
}
