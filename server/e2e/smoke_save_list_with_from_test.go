package e2e_test

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/e2e"
	"github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Confirm logic around lists with from object in the info block", func() {
	var (
		ListWithNoFrom                  = "./list-from-001.json"
		ListWithFrom                    = "./list-from-with-from.json"
		ListWithFromAndKindNotSupported = "./list-from-with-from-kind-not-supported.json"
		ListWithFromAndKindIsLearnalist = "./list-from-with-from-kind-learnalist.json"
	)

	var (
		openapiClient  e2e.OpenApiClient
		e2eClient      e2e.Client
		userInfoOwnerA e2e.RegisterResponse
		userInfoOwnerB e2e.RegisterResponse
		password       = "test123"
		authA          context.Context
		authB          context.Context
	)

	BeforeEach(func() {
		openapiClient = e2e.NewOpenApiClient(e2e.LOCAL_SERVER)
		e2eClient = e2e.NewClient("http://localhost:1234")
		userInfoOwnerA = e2eClient.Register(generateUsername(), password)

		authA = context.WithValue(context.Background(), openapi.ContextBasicAuth, openapi.BasicAuth{
			UserName: userInfoOwnerA.Username,
			Password: password,
		})

		userInfoOwnerB = e2eClient.Register(generateUsername(), password)

		authB = context.WithValue(context.Background(), openapi.ContextBasicAuth, openapi.BasicAuth{
			UserName: userInfoOwnerB.Username,
			Password: password,
		})
	})

	AfterEach(func() {
		openapiClient.DeleteUser(authA, userInfoOwnerA.Uuid)
		openapiClient.DeleteUser(authB, userInfoOwnerB.Uuid)
	})

	When("Not present", func() {
		It("Should be normal behaviour", func() {
			_, err := e2eClient.PostListV1(userInfoOwnerA, testutils.GetTestDataAsJSONOneline(ListWithNoFrom))
			Expect(err).To(BeNil())
		})
	})

	When("from is present in the info block", func() {
		When("Saving a new list", func() {
			It("from kind is allowed", func() {
				aList, err := e2eClient.PostListV1(userInfoOwnerA, testutils.GetTestDataAsJSONOneline(ListWithFrom))
				Expect(err).To(BeNil())
				Expect(aList.Info.From.Kind).To(Equal("quizlet"))

				By("Confirm handling when refurl doesnt match the kind")
				aList.Info.From.RefUrl = "https://notreal.com/xxx"
				data, _ := json.Marshal(aList)
				resp, err := testutils.ToHttpResponse(e2eClient.RawPostListV1(userInfoOwnerA, string(data)))
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusUnprocessableEntity))
				testutils.CheckMessageResponse(resp, i18n.ErrorAListFromDomainMisMatch.Error())
			})

			Specify("fail when trying to save with shared set to anything but private", func() {
				aList, err := e2eClient.PostListV1(userInfoOwnerA, testutils.GetTestDataAsJSONOneline(ListWithFrom))
				Expect(err).To(BeNil())
				aList.Uuid = ""
				aList.Info.SharedWith = keys.SharedWithPublic
				data, _ := json.Marshal(aList)
				resp, err := testutils.ToHttpResponse(e2eClient.RawPostListV1(userInfoOwnerA, string(data)))
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusUnprocessableEntity))
				testutils.CheckMessageResponse(resp, i18n.InputSaveAlistOperationFromRestriction)
			})

			It("from kind is not allowed", func() {
				resp, err := testutils.ToHttpResponse(e2eClient.RawPostListV1(userInfoOwnerA, testutils.GetTestDataAsJSONOneline(ListWithFromAndKindNotSupported)))
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusUnprocessableEntity))
				testutils.CheckMessageResponse(resp, i18n.InputSaveAlistOperationFromKindNotSupported)
			})
		})

		When("Updating a list", func() {
			It("Make sure the from block cant be modified", func() {
				aList, err := e2eClient.PostListV1(userInfoOwnerA, testutils.GetTestDataAsJSONOneline(ListWithFrom))
				Expect(err).To(BeNil())
				aList.Info.From.Kind = "cram"
				data, _ := json.Marshal(aList)
				resp, err := testutils.ToHttpResponse(e2eClient.RawPutListV1(userInfoOwnerA, aList.Uuid, string(data)))
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusUnprocessableEntity))
				testutils.CheckMessageResponse(resp, i18n.ErrorAListFromDomainMisMatch.Error())
			})

			When("from, not learnalist", func() {
				It("Make sure shared is not modified", func() {
					aList, err := e2eClient.PostListV1(userInfoOwnerA, testutils.GetTestDataAsJSONOneline(ListWithFrom))
					Expect(err).To(BeNil())
					aList.Info.SharedWith = keys.SharedWithFriends
					data, _ := json.Marshal(aList)
					resp, err := testutils.ToHttpResponse(e2eClient.RawPutListV1(userInfoOwnerA, aList.Uuid, string(data)))
					Expect(err).To(BeNil())
					Expect(resp.StatusCode).To(Equal(http.StatusUnprocessableEntity))
					testutils.CheckMessageResponse(resp, i18n.ErrorInputSaveAlistOperationFromRestriction.Error())
				})
			})

			When("from learnalist", func() {
				It("Shared can be changed", func() {
					aList, err := e2eClient.PostListV1(userInfoOwnerA, testutils.GetTestDataAsJSONOneline(ListWithFromAndKindIsLearnalist))
					Expect(err).To(BeNil())
					aList.Info.SharedWith = keys.SharedWithFriends
					data, _ := json.Marshal(aList)
					resp, err := testutils.ToHttpResponse(e2eClient.RawPutListV1(userInfoOwnerA, aList.Uuid, string(data)))
					Expect(err).To(BeNil())
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
				})
			})
		})
	})
})
