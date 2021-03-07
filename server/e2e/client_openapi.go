package e2e

import (
	"context"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	. "github.com/onsi/gomega"
)

type OpenApiClient struct {
	Config *openapi.Configuration
	API    *openapi.APIClient
	logs   []HTTPLog
}

type HTTPLog struct {
	Method     string `json:"method"`
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
}

const (
	LOCAL_SERVER = "http://localhost:1234/api/v1"
)

func NewOpenApiClient(server string) *OpenApiClient {
	oasClient := OpenApiClient{}

	config := openapi.NewConfiguration()
	config.BasePath = server
	client := openapi.NewAPIClient(config)
	oasClient.API = client
	oasClient.Config = config

	httpClient := http.Client{}
	config.HTTPClient.Transport = testutils.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		// If we change the rule?
		// How could I add the operationID as a header (would make things much simpler)

		resp, err := httpClient.Do(req)
		logEntry := HTTPLog{
			Method:     req.Method,
			URL:        req.URL.Path, // TODO does this include query
			StatusCode: resp.StatusCode,
		}
		oasClient.AddLog(logEntry)
		return resp, err
	})

	return &oasClient
}

func (c *OpenApiClient) AddLog(l HTTPLog) {
	c.logs = append(c.logs, l)
}

func (c *OpenApiClient) GetLogs() []HTTPLog {
	return c.logs
}

func (c OpenApiClient) RegisterUser(input openapi.HttpUserRegisterInput) openapi.HttpUserRegisterResponse {
	data, response, err := c.API.UserApi.RegisterUserWithUsernameAndPassword(context.Background(), input, nil)

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
