package e2e_test

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/pkg/apps"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing remind API", func() {
	var client *openapi.APIClient

	BeforeEach(func() {
		config := openapi.NewConfiguration()
		config.BasePath = "http://localhost:1234/api/v1"
		client = openapi.NewAPIClient(config)
	})

	When("Saving daily settings", func() {
		It("Happy Path", func() {
			// TODO this is going to be replace in the future.
			// Register User
			auth, loginInfo := RegisterAndLogin(client)
			input := openapi.RemindDailySettings{
				TimeOfDay:     "00:00",
				Tz:            "Europe/Oslo",
				AppIdentifier: apps.RemindV1,
				Medium:        []string{"push"},
			}

			settings, resp, err := client.RemindApi.SetRemindDailySetting(auth, input)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Expect(settings.AppIdentifier).To(Equal(input.AppIdentifier))
			Expect(settings.Tz).To(Equal(input.Tz))
			Expect(settings.TimeOfDay).To(Equal(input.TimeOfDay))
			Expect(settings.Medium).To(Equal(input.Medium))

			// Delete user
			DeleteUser(client, auth, loginInfo.UserUuid)
		})
	})
	When("GetDailySettings", func() {
		It("App not supported", func() {
			// Register User
			auth, loginInfo := RegisterAndLogin(client)

			_, resp, err := client.RemindApi.GetRemindDailySettingsByAppIdentifier(auth, "fake-app")
			Expect(err).To(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusUnprocessableEntity))
			msg := err.(openapi.GenericOpenAPIError).Model().(openapi.HttpResponseMessage).Message
			Expect(msg).To(Equal("appIdentifier is not valid"))
			// Delete user
			DeleteUser(client, auth, loginInfo.UserUuid)
		})

		It("No settings", func() {
			// Register User
			auth, loginInfo := RegisterAndLogin(client)

			_, resp, err := client.RemindApi.GetRemindDailySettingsByAppIdentifier(auth, apps.RemindV1)

			Expect(err).To(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			msg := err.(openapi.GenericOpenAPIError).Model().(openapi.HttpResponseMessage).Message
			Expect(msg).To(Equal("Settings not found"))
			// Delete user
			DeleteUser(client, auth, loginInfo.UserUuid)
		})
	})

})
