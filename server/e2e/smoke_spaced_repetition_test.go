package e2e_test

import (
	"encoding/json"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Spaced repetition API", func() {
	var client *openapi.APIClient

	BeforeEach(func() {
		// TODO use global
		config := openapi.NewConfiguration()
		config.BasePath = "http://localhost:1234/api/v1"
		client = openapi.NewAPIClient(config)
	})

	/*
		srs := server.Group("/api/v1/spaced-repetition")
		srs.Use(authenticate.Auth(authConfig))
		srs.GET("/next", spacedRepetitionService.GetNext)
		srs.GET("/all", spacedRepetitionService.GetAll)
		srs.DELETE("/:uuid", spacedRepetitionService.DeleteEntry)
		srs.POST("/", spacedRepetitionService.SaveEntry)
		srs.POST("/viewed", spacedRepetitionService.EntryViewed)
	*/
	When("Adding new entry for learning", func() {
		It("Bug GH-185: Two users post the same entry", func() {
			// Add entry via two users, confirm they add
			// Confirm re-adding = 200
			var (
				b     []byte
				input map[string]interface{}
				entry openapi.SpacedRepetitionV1
			)
			auth1, loginInfo1 := RegisterAndLogin(client)
			auth2, loginInfo2 := RegisterAndLogin(client)

			body := openapi.SpacedRepetitionV1{
				Show: "Hello",
				Data: "Hello",
				Kind: alist.SimpleList,
			}
			b, _ = json.Marshal(body)
			json.Unmarshal(b, &input)

			entryResp, resp, err := client.SpacedRepetitionApi.AddSpacedRepetitionEntry(auth1, input)
			Expect(err).To(BeNil())
			b, _ = json.Marshal(entryResp)

			json.Unmarshal(b, &entry)
			Expect(entry.Uuid).To(Equal("ba9277fc4c6190fb875ad8f9cee848dba699937f"))
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			entryResp, resp, err = client.SpacedRepetitionApi.AddSpacedRepetitionEntry(auth2, input)
			Expect(err).To(BeNil())
			b, _ = json.Marshal(entryResp)

			json.Unmarshal(b, &entry)
			Expect(entry.Uuid).To(Equal("ba9277fc4c6190fb875ad8f9cee848dba699937f"))
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			_, resp, err = client.SpacedRepetitionApi.AddSpacedRepetitionEntry(auth2, input)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Delete user
			DeleteUser(client, auth1, loginInfo1.UserUuid)
			DeleteUser(client, auth2, loginInfo2.UserUuid)
		})
	})

})
