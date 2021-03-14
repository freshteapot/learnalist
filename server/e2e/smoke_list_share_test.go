package e2e_test

import (
	"fmt"

	"github.com/freshteapot/learnalist-api/server/e2e"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing API share a list", func() {
	var (
		userInfoOwnerA e2e.RegisterResponse
	)

	It("list sharing2", func() {
		Expect("").To(Equal(""))
		password := "test123"

		userInfoOwnerA = e2eClient.Register(generateUsername(), password)

		//authA = context.WithValue(context.Background(), openapi.ContextAccessToken, openapi.BasicAuth{
		//	UserName: userInfoOwnerA.Username,
		//	Password: password,
		//})

		authOwner, loginInfoOwner := RegisterAndLogin(openapiClient.API)
		fmt.Println(userInfoOwnerA, loginInfoOwner, authOwner.Value(openapi.ContextAccessToken).(string))
		//input := openapi.HttpUserRegisterInput{
		//	Username: generateUsername(),
		//	Password: "test123",
		//}
		//fmt.Println(authOwner)
		//fmt.Println(loginInfoOwner)
		//
		//resp, err := e2e.RawPostListV1(openapiClient.Config.BasePath, authOwner, testutils.GetTestDataAsJSONOneline(ListWithNoFrom))
		////json.Unmarshal(testutils.GetTestData(ListWithNoFrom), &aListInput)
		////aList, resp, err := openapiClient.API.AListApi.AddList(authOwner, aListInput)
		//fmt.Println(resp, err)

		// Delete user
		//DeleteUser(client, authOwner, loginInfoOwner.UserUuid)
	})
})
