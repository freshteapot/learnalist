package e2e_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Testing openapi", func() {

	It("Get version", func() {
		config := openapi.NewConfiguration()
		config.BasePath = "http://localhost:1234/api/v1"
		client := openapi.NewAPIClient(config)

		version, response, _ := client.DefaultApi.GetServerVersion(context.Background())
		fmt.Println(version)
		Expect(response.StatusCode).To(Equal(http.StatusOK))
	})

	It("Get Next", func() {
		config := openapi.NewConfiguration()
		config.BasePath = "http://localhost:1234/api/v1"

		auth := context.WithValue(context.Background(), openapi.ContextBasicAuth, openapi.BasicAuth{
			UserName: "iamchris",
			Password: "test123",
		})

		client := openapi.NewAPIClient(config)
		data, response, err := client.SpacedRepetitionApi.GetNextSpacedRepetitionEntry(auth)
		Expect(err).Should(HaveOccurred())
		Expect(response.StatusCode).To(Equal(http.StatusNotFound))
		fmt.Println(data)
	})

	It("Register user", func() {
		config := openapi.NewConfiguration()
		config.BasePath = "http://localhost:1234/api/v1"
		client := openapi.NewAPIClient(config)
		input := openapi.HttpUserRegisterInput{
			Username: "iamchris1",
			Password: "test123",
		}

		data1, response, err := client.UserApi.RegisterUserWithUsernameAndPassword(context.Background(), input)
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusCreated))
		Expect(data1.Username).To(Equal(input.Username))

		data2, response, err := client.UserApi.RegisterUserWithUsernameAndPassword(context.Background(), input)
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
		Expect(data2.Username).To(Equal(input.Username))
		Expect(data2.Uuid).To(Equal(data1.Uuid))

		auth := context.WithValue(context.Background(), openapi.ContextBasicAuth, openapi.BasicAuth{
			UserName: input.Username,
			Password: input.Password,
		})

		data, response, err := client.UserApi.DeleteUser(auth, data1.Uuid)
		fmt.Println(data, response, err)
	})
})
