package e2e_test

import (
	"context"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	. "github.com/onsi/gomega"
)

func RegisterAndLogin(client *openapi.APIClient) (context.Context, openapi.HttpUserLoginResponse) {
	ctx := context.Background()
	input := openapi.HttpUserRegisterInput{
		Username: generateUsername(),
		Password: "test123",
	}
	data1, response, err := client.UserApi.RegisterUserWithUsernameAndPassword(ctx, input, nil)
	Expect(err).To(BeNil())
	Expect(response.StatusCode).To(Equal(http.StatusCreated))
	Expect(data1.Username).To(Equal(input.Username))

	// Login
	loginInfo, response, err := client.UserApi.LoginWithUsernameAndPassword(ctx, openapi.HttpUserLoginRequest{
		Username: input.Username,
		Password: input.Password,
	})
	Expect(err).To(BeNil())
	Expect(response.StatusCode).To(Equal(http.StatusOK))
	Expect(loginInfo.UserUuid).To(Equal(data1.Uuid))
	auth := context.WithValue(ctx, openapi.ContextAccessToken, loginInfo.Token)
	return auth, loginInfo
}

func DeleteUser(client *openapi.APIClient, auth context.Context, userUUID string) {
	_, response, err := client.UserApi.DeleteUser(auth, userUUID)
	Expect(err).To(BeNil())
	Expect(response.StatusCode).To(Equal(http.StatusOK))
}
