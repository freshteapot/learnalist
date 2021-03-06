package e2e_test

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/pkg/apps"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing mobile API", func() {
	var client *openapi.APIClient

	BeforeEach(func() {
		config := openapi.NewConfiguration()
		config.BasePath = "http://localhost:1234/api/v1"
		client = openapi.NewAPIClient(config)
	})

	When("Register new device", func() {
		It("Currently supported app plank:v1", func() {
			// TODO this is going to be replace in the future.
			// Register User
			auth, loginInfo := RegisterAndLogin(client)

			// Register Device
			deviceInput := openapi.HttpMobileRegisterInput{
				Token:         "fake-token-123",
				AppIdentifier: "plank:v1",
			}
			msg, response, err := client.MobileApi.RegisterDevice(auth, deviceInput)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			Expect(msg.Message).To(Equal("Device registered"))

			// Delete user
			DeleteUser(client, auth, loginInfo.UserUuid)
		})

		It("Fail on invalid app", func() {
			// Register User
			auth, loginInfo := RegisterAndLogin(client)

			// Register Device
			deviceInput := openapi.HttpMobileRegisterInput{
				Token:         "fake-token-123",
				AppIdentifier: "fake-app",
			}
			_, response, err := client.MobileApi.RegisterDevice(auth, deviceInput)

			Expect(err).To(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusUnprocessableEntity))
			msg := err.(openapi.GenericOpenAPIError).Model().(openapi.HttpResponseMessage).Message
			Expect(msg).To(Equal("App identifier is not supported: plank_v1,remind_v1"))
			// Delete user
			DeleteUser(client, auth, loginInfo.UserUuid)
		})

		It("Register single device", func() {
			// Register User
			auth, loginInfo := RegisterAndLogin(client)

			// Register Device
			deviceInput := openapi.HttpMobileRegisterInput{
				Token:         "fake-token-123",
				AppIdentifier: apps.RemindV1,
			}
			msg, response, err := client.MobileApi.RegisterDevice(auth, deviceInput)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			Expect(msg.Message).To(Equal("Device registered"))

			// Delete user
			DeleteUser(client, auth, loginInfo.UserUuid)
		})

		It("Register device twice", func() {
			// Register User
			auth, loginInfo := RegisterAndLogin(client)

			// Register Device
			deviceInput := openapi.HttpMobileRegisterInput{
				Token:         "fake-token-123",
				AppIdentifier: apps.RemindV1,
			}
			msg, response, err := client.MobileApi.RegisterDevice(auth, deviceInput)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			Expect(msg.Message).To(Equal("Device registered"))

			msg, response, err = client.MobileApi.RegisterDevice(auth, deviceInput)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			Expect(msg.Message).To(Equal("Device registered"))

			// Delete user
			DeleteUser(client, auth, loginInfo.UserUuid)
		})

		It("Register device and replace it", func() {
			// Register User
			auth1, loginInfo1 := RegisterAndLogin(client)
			auth2, loginInfo2 := RegisterAndLogin(client)

			// Register Device
			deviceInput := openapi.HttpMobileRegisterInput{
				Token:         "fake-token-123",
				AppIdentifier: apps.RemindV1,
			}
			msg, response, err := client.MobileApi.RegisterDevice(auth1, deviceInput)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			Expect(msg.Message).To(Equal("Device registered"))

			msg, response, err = client.MobileApi.RegisterDevice(auth2, deviceInput)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			Expect(msg.Message).To(Equal("Device registered"))

			// Delete user
			DeleteUser(client, auth1, loginInfo1.UserUuid)
			DeleteUser(client, auth2, loginInfo2.UserUuid)
		})
	})

})
