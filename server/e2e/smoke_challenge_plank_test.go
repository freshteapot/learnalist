package e2e_test

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/antihax/optional"
	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Smoke Challenge Plank", func() {

	It("Test 1", func() {
		// Create user
		// Login
		// Create challenge
		// Add plank
		// Delete user

		// Create user
		ctx := context.Background()
		input := openapi.HttpUserRegisterInput{
			Username: generateUsername(),
			Password: "test123",
		}
		data1, response, err := openapiClient.API.UserApi.RegisterUserWithUsernameAndPassword(ctx, input, nil)
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusCreated))
		Expect(data1.Username).To(Equal(input.Username))

		// Login
		loginInfo, response, err := openapiClient.API.UserApi.LoginWithUsernameAndPassword(ctx, openapi.HttpUserLoginRequest{
			Username: input.Username,
			Password: input.Password,
		})
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
		Expect(loginInfo.UserUuid).To(Equal(data1.Uuid))
		auth := context.WithValue(ctx, openapi.ContextAccessToken, loginInfo.Token)

		// Create challenge
		challengeInput := openapi.ChallengeInput{
			Description: "hello",
			Kind:        challenge.KindPlankGroup,
		}
		info, response, err := openapiClient.API.ChallengeApi.CreateChallenge(auth, challengeInput)
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusCreated))
		Expect(info.Kind).To(Equal(challengeInput.Kind))
		Expect(info.Description).To(Equal(challengeInput.Description))

		// Add plank
		record1JSON := `
		{
			"showIntervals": true,
			"intervalTime": 15,
			"beginningTime": 1602264153548,
			"currentTime": 1602264219291,
			"timerNow": 65743,
			"intervalTimerNow": 5681,
			"laps": 4
		}`
		var record1 openapi.Plank

		json.Unmarshal([]byte(record1JSON), &record1)
		_, response, err = openapiClient.API.PlankApi.AddPlankEntry(auth, record1, &openapi.AddPlankEntryOpts{XChallenge: optional.NewString(info.Uuid)})
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusCreated))
		// Double check posting the same plank returns 200
		_, response, err = openapiClient.API.PlankApi.AddPlankEntry(auth, record1, &openapi.AddPlankEntryOpts{XChallenge: optional.NewString(info.Uuid)})
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
		// Add another record
		record2JSON := `
		{
			"showIntervals": false,
			"intervalTime": 0,
			"beginningTime": 1602264153549,
			"currentTime": 1602264219291,
			"timerNow": 65742,
			"intervalTimerNow": 65742,
			"laps": 0
		}`
		var record2 openapi.Plank

		json.Unmarshal([]byte(record2JSON), &record2)
		record2, response, err = openapiClient.API.PlankApi.AddPlankEntry(auth, record2, &openapi.AddPlankEntryOpts{XChallenge: optional.NewString(info.Uuid)})
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusCreated))

		challengeInfo, response, err := openapiClient.API.ChallengeApi.GetChallenge(auth, info.Uuid)

		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
		// Confirm record 2 is latest, as it was the latest event
		Expect(len(challengeInfo.Records)).To(Equal(2))
		a, _ := json.Marshal(challengeInfo.Records[0])
		var expect openapi.Plank
		json.Unmarshal(a, &expect)
		Expect(expect).To(Equal(record2))

		// Delete user
		_, response, err = openapiClient.API.UserApi.DeleteUser(auth, data1.Uuid)
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
	})

	It("Test filtering via kind", func() {
		// Create user
		// Login
		// Create challenge TODO
		// Create challenge plank-group
		// Query all
		// Query TODO
		// Query plank-group
		// Delete user

		// Create user
		ctx := context.Background()
		input := openapi.HttpUserRegisterInput{
			Username: generateUsername(),
			Password: "test123",
		}
		data1, response, err := openapiClient.API.UserApi.RegisterUserWithUsernameAndPassword(ctx, input, nil)
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusCreated))
		Expect(data1.Username).To(Equal(input.Username))

		// Login
		loginInfo, response, err := openapiClient.API.UserApi.LoginWithUsernameAndPassword(ctx, openapi.HttpUserLoginRequest{
			Username: input.Username,
			Password: input.Password,
		})
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
		Expect(loginInfo.UserUuid).To(Equal(data1.Uuid))
		auth := context.WithValue(ctx, openapi.ContextAccessToken, loginInfo.Token)

		// Create challenge TODO
		challengeInput := openapi.ChallengeInput{
			Description: "hello",
			Kind:        challenge.KindTODO,
		}
		info, response, err := openapiClient.API.ChallengeApi.CreateChallenge(auth, challengeInput)
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusCreated))
		Expect(info.Kind).To(Equal(challengeInput.Kind))
		Expect(info.Description).To(Equal(challengeInput.Description))

		// Create challenge plankGroup
		challengeInput = openapi.ChallengeInput{
			Description: "hello",
			Kind:        challenge.KindPlankGroup,
		}
		info, response, err = openapiClient.API.ChallengeApi.CreateChallenge(auth, challengeInput)
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusCreated))
		Expect(info.Kind).To(Equal(challengeInput.Kind))
		Expect(info.Description).To(Equal(challengeInput.Description))

		items, response, err := openapiClient.API.ChallengeApi.GetChallengesByUser(auth, loginInfo.UserUuid, nil)
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
		Expect(len(items)).To(Equal(2))

		items, response, err = openapiClient.API.ChallengeApi.GetChallengesByUser(auth, loginInfo.UserUuid, &openapi.GetChallengesByUserOpts{Kind: optional.NewInterface(challenge.KindPlankGroup)})
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
		Expect(len(items)).To(Equal(1))

		// Delete user
		_, response, err = openapiClient.API.UserApi.DeleteUser(auth, data1.Uuid)
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
	})
})
