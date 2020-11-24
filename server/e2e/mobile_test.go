package e2e_test

import (
	"context"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Smoke Mobile", func() {

	var client *openapi.APIClient

	BeforeEach(func() {
		config := openapi.NewConfiguration()
		config.BasePath = "http://localhost:1234/api/v1"
		client = openapi.NewAPIClient(config)
	})

	It("Confirm one device even if added twice", func() {
		// Create user
		// Register device
		// Register device x2
		// Manually check the db

		// Create user
		ctx := context.Background()
		input := openapi.HttpUserRegisterInput{
			Username: generateUsername(),
			Password: "test123",
		}
		data1, response, err := client.UserApi.RegisterUserWithUsernameAndPassword(ctx, input)
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

		// Register device
		mobileInput := openapi.HttpMobileRegisterInput{
			Token: "fake-token-1",
		}

		_, response, err = client.MobileApi.RegisterDevice(auth, mobileInput)
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusOK))

		_, response, err = client.MobileApi.RegisterDevice(auth, mobileInput)
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusOK))

		// Delete user
		_, response, err = client.UserApi.DeleteUser(auth, data1.Uuid)
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
	})
})
