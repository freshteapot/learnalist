package api

import (
	"fmt"

	"github.com/freshteapot/learnalist-api/server/e2e"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"

	"github.com/spf13/cobra"
)

var deleteUserCmd = &cobra.Command{
	Use:   "delete-user",
	Short: "Remove a user from the system",
	Run: func(cmd *cobra.Command, args []string) {
		//logger := logging.GetLogger()
		server := "http://127.0.0.1:1234"

		credentials := openapi.HttpUserLoginRequest{
			Username: "iamtest1",
			Password: "test123",
		}

		learnalistClient := e2e.NewClient(server)
		statusCode, response := learnalistClient.Login(credentials)
		fmt.Println(statusCode)
		fmt.Println(response)

		deleteStatusCode, deleteResponse := learnalistClient.DeleteUser(response)
		fmt.Println(deleteStatusCode)
		fmt.Println(deleteResponse)
	},
}
