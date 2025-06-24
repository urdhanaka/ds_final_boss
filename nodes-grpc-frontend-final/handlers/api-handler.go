package handlers

import (
	"fmt"
	"nodes-grpc-fe/consts"
	"nodes-grpc-fe/models"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
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

func (api *ApiClient) Login(c *gin.Context, email, password string) (*models.ApiResponse, error) {
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

func (api *ApiClient) Me(c *gin.Context) (*models.ApiResponse, error) {
	token := c.MustGet("token").(string)

	bearerToken := fmt.Sprintf("Bearer %s", token)

	resp, err := api.client.R().
		SetHeader("Authorization", bearerToken).
		SetResult(&models.ApiResponse{}).
		Get(consts.BACKEND_URL + "/api/users/me")
	if err != nil {
		return &models.ApiResponse{}, err
	}

	respModel := resp.Result().(*models.ApiResponse)

	return respModel, nil
}
