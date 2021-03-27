package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/freshteapot/learnalist-api/server/pkg/apps"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/nats-io/stan.go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Regression GH-233", func() {
	It("Confirming that 2 different users get daily notifications, instead of 1 user with many tokens", func() {
		client := openapiClient.API
		authA, loginInfoA := RegisterAndLogin(client)
		authB, loginInfoB := RegisterAndLogin(client)
		// if this allowed seconds, then the test wouldnt have to wait
		future := time.Now().Add(1 * time.Second)
		input := openapi.RemindDailySettings{
			TimeOfDay:     fmt.Sprintf("%02d:%02d:%02d", future.Hour(), future.Minute(), future.Second()),
			Tz:            "Europe/Oslo",
			AppIdentifier: apps.RemindV1,
			Medium:        []string{"push"},
		}
		// Register Devices
		deviceInput := openapi.HttpMobileRegisterInput{
			Token:         "fake-token-123",
			AppIdentifier: apps.RemindV1,
		}
		msg, response, err := client.MobileApi.RegisterDevice(authA, deviceInput)

		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
		Expect(msg.Message).To(Equal("Device registered"))

		deviceInput.Token = "fake-token-123-b"
		msg, response, err = client.MobileApi.RegisterDevice(authB, deviceInput)

		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
		Expect(msg.Message).To(Equal("Device registered"))

		_, _, err = client.RemindApi.SetRemindDailySetting(authA, input)
		Expect(err).To(BeNil())
		_, _, err = client.RemindApi.SetRemindDailySetting(authB, input)
		Expect(err).To(BeNil())

		sc := event.GetBus().(*event.NatsBus).Connection()
		// A little ugly, but a few seconds more than the daily_manager.StartSendNotifications
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)

		found := map[string]bool{}
		found[loginInfoA.UserUuid] = false
		found[loginInfoB.UserUuid] = false

		handle := func(msg *stan.Msg) {
			var moment event.Eventlog
			err := json.Unmarshal(msg.Data, &moment)
			Expect(err).To(BeNil())
			found[moment.UUID] = true

			if found[loginInfoA.UserUuid] && found[loginInfoB.UserUuid] {
				cancel()
			}
		}
		subscription, _ := sc.Subscribe(
			event.TopicNotifications,
			handle,
			stan.MaxInflight(1),
		)

		select {
		case <-ctx.Done():
			Expect(found[loginInfoA.UserUuid] && found[loginInfoB.UserUuid]).To(BeTrue(), "Both users should have gotten a notification")
		}

		subscription.Close()

		// Delete users
		DeleteUser(client, authA, loginInfoA.UserUuid)
		DeleteUser(client, authB, loginInfoB.UserUuid)
	})
})
