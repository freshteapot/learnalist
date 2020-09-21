package e2e

import (
	"context"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	. "github.com/onsi/gomega"
)

type OpenApiClient struct {
	Config *openapi.Configuration
	API    *openapi.APIClient
}

const (
	LOCAL_SERVER = "http://localhost:1234/api/v1"
)

func NewOpenApiClient(server string) OpenApiClient {
	config := openapi.NewConfiguration()
	config.BasePath = server
	client := openapi.NewAPIClient(config)

	return OpenApiClient{
		API:    client,
		Config: config,
	}
}

func (c OpenApiClient) RegisterUser(input openapi.HttpUserRegisterInput) openapi.HttpUserRegisterResponse {
	data, response, err := c.API.UserApi.RegisterUserWithUsernameAndPassword(context.Background(), input)

	Expect(err).To(BeNil())
	Expect(response.StatusCode).To(Equal(http.StatusCreated))
	Expect(data.Username).To(Equal(input.Username))
	return data
}

func (c OpenApiClient) DeleteUser(ctx context.Context, uuid string) {
	_, response, err := c.API.UserApi.DeleteUser(ctx, uuid)
	Expect(err).To(BeNil())
	Expect(response.StatusCode).To(Equal(http.StatusOK))
}
