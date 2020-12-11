package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/antihax/optional"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing openapi", func() {
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
			UserName: "iamtest1",
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
			Username: "iamtest11",
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

	// go clean -testcache && go test --tags="json1" -ginkgo.v -ginkgo.progress -ginkgo.focus="Testing openapi Upload asset" -test.v .
	It("Upload asset", func() {
		config := openapi.NewConfiguration()
		config.BasePath = "http://localhost:1234/api/v1"
		client := openapi.NewAPIClient(config)
		input := openapi.HttpUserRegisterInput{
			Username: "iamtest12",
			Password: "test123",
		}

		auth := context.WithValue(context.Background(), openapi.ContextBasicAuth, openapi.BasicAuth{
			UserName: input.Username,
			Password: input.Password,
		})

		user, _, err := client.UserApi.RegisterUserWithUsernameAndPassword(context.Background(), input)
		Expect(err).To(BeNil())

		uploadFile, _ := os.Open("./testdata/sample.png")
		defer uploadFile.Close()

		opts := openapi.AddUserAssetOpts{
			SharedWith: optional.NewString("public"),
		}
		asset, response, err := client.AssetApi.AddUserAsset(auth, uploadFile, &opts)
		Expect(err).To(BeNil())
		Expect(response.StatusCode).To(Equal(http.StatusCreated))
		Expect(asset.Href).ToNot(BeEmpty())
		Expect(asset.Href).To(ContainSubstring(user.Uuid))

		// Test public and private list
		// Not sure how I feel about this
		assetConfig := openapi.NewConfiguration()
		assetConfig.BasePath = "http://localhost:1234"
		assetClient := openapi.NewAPIClient(assetConfig)
		assetUUID := strings.TrimPrefix(asset.Href, "/assets/")
		response, _ = assetClient.AssetApi.GetAsset(context.Background(), assetUUID)
		Expect(response.StatusCode).To(Equal(http.StatusOK))

		_, response, _ = client.AssetApi.ShareAsset(auth, openapi.HttpAssetShareRequestBody{Uuid: asset.Uuid, Action: aclKeys.NotShared})
		Expect(response.StatusCode).To(Equal(http.StatusOK))

		response, _ = assetClient.AssetApi.GetAsset(context.Background(), assetUUID)
		Expect(response.StatusCode).To(Equal(http.StatusForbidden))

		_, _, err = client.UserApi.DeleteUser(auth, user.Uuid)
		Expect(err).To(BeNil())
	})
})
