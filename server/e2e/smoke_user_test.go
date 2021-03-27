package e2e_test

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/pkg/apps"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Smoke user", func() {
	var (
		loginInfo openapi.HttpUserLoginResponse
		auth      context.Context
		pref      user.UserPreference
	)

	AfterEach(func() {
		if loginInfo.UserUuid != "" {
			openapiClient.DeleteUser(auth, loginInfo.UserUuid)
		}
	})

	//
	It("Register, login and delete", func() {
		// Register
		username := generateUsername()
		e2eClient.Register(username, password)

		// Login
		response, err := e2eClient.RawLogin(username, password)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
		defer response.Body.Close()

		decoder := json.NewDecoder(response.Body)
		var userCredentials openapi.HttpUserLoginResponse
		err = decoder.Decode(&userCredentials)
		Expect(err).ShouldNot(HaveOccurred())

		// Delete
		statusCode, httpResponse := e2eClient.DeleteUser(userCredentials)
		Expect(statusCode).To(Equal(http.StatusOK))
		Expect(convertResponseToString(httpResponse)).To(Equal(`{"message":"User has been removed"}`))
	})

	It("Changing display name", func() {
		auth, loginInfo = RegisterAndLogin(openapiClient.API)
		raw, resp, err := openapiClient.API.UserApi.GetUserInfo(auth, loginInfo.UserUuid)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		pref = user.UserPreference{}
		testutils.ConvertInterface(raw, &pref)
		Expect(pref.Acl.PublicListWrite).To(Equal(1))
		Expect(pref.UserUUID).To(Equal(loginInfo.UserUuid))
		Expect(pref.DisplayName).To(Equal("Chris"))

		resp, err = openapiClient.API.UserApi.PatchUserInfo(auth, loginInfo.UserUuid, openapi.HttpUserInfoInput{
			DisplayName: "Bob",
		})
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		raw, resp, err = openapiClient.API.UserApi.GetUserInfo(auth, loginInfo.UserUuid)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		pref = user.UserPreference{}
		testutils.ConvertInterface(raw, &pref)
		Expect(pref.DisplayName).To(Equal("Bob"))
	})

	It("Confirm what happens after we update the daily reminder settings", func() {
		// Register
		auth, loginInfo = RegisterAndLogin(openapiClient.API)
		raw, resp, err := openapiClient.API.UserApi.GetUserInfo(auth, loginInfo.UserUuid)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		pref = user.UserPreference{}
		testutils.ConvertInterface(raw, &pref)
		Expect(pref.Acl.PublicListWrite).To(Equal(1))
		Expect(pref.UserUUID).To(Equal(loginInfo.UserUuid))
		Expect(pref.DisplayName).To(Equal("Chris"))
		// Add DailyReminderSettings
		input := openapi.RemindDailySettings{
			TimeOfDay:     "01:00:00",
			Tz:            "Europe/Oslo",
			AppIdentifier: apps.RemindV1,
			Medium:        []string{"push"},
		}
		_, resp, err = openapiClient.API.RemindApi.SetRemindDailySetting(auth, input)

		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		raw, resp, err = openapiClient.API.UserApi.GetUserInfo(auth, loginInfo.UserUuid)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		pref = user.UserPreference{}
		testutils.ConvertInterface(raw, &pref)
		Expect(pref.DailyReminder.RemindV1.TimeOfDay).To(Equal(input.TimeOfDay))

		_, resp, err = openapiClient.API.RemindApi.DeleteRemindDailySettingsByAppIdentifier(auth, apps.RemindV1)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		raw, resp, err = openapiClient.API.UserApi.GetUserInfo(auth, loginInfo.UserUuid)
		Expect(err).To(BeNil())
		pref = user.UserPreference{}
		testutils.ConvertInterface(raw, pref)
		Expect(pref.DailyReminder).To(BeNil())

	})
})
