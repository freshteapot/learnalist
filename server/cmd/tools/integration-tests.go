package tools

import (
	"fmt"
	"os"

	"github.com/freshteapot/learnalist-api/server/api/client"
	"github.com/freshteapot/learnalist-api/server/api/integrations"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var integrationTestsCmd = &cobra.Command{
	Use:   "integration-tests",
	Short: "Run integration tests against a learnalist server",
	Run: func(cmd *cobra.Command, args []string) {
		server := viper.GetString("tools.integrationTests.server")
		username := viper.GetString("tools.integrationTests.username")
		password := viper.GetString("tools.integrationTests.password")
		runTests := viper.GetBool("tools.integrationTests.runTests")
		showSupported := viper.GetBool("tools.integrationTests.showSupported")

		if showSupported {
			supported()
			os.Exit(0)
		}

		config := client.Config{
			Server:   server,
			Username: username,
			Password: password,
		}

		client := client.Client{
			Config: config,
		}

		if runTests {
			integrations := integrations.Client{
				ApiClient: client,
			}

			integrations.RunIntegrationTests()
			os.Exit(0)
		}

		rootResponse, _ := client.GetRoot()
		fmt.Println(rootResponse)
		versionResponse, _ := client.GetVersion()
		fmt.Println(versionResponse)
	},
}

func init() {
	integrationTestsCmd.Flags().Bool("show-supported", false, "When set, show the api endpoints supported by the client")
	integrationTestsCmd.Flags().Bool("run", false, "Run the integration tests against the server")
	integrationTestsCmd.Flags().String("server", "https://learnalist.net/api/v1", "The server.")
	viper.BindPFlag("tools.integrationTests.server", integrationTestsCmd.Flags().Lookup("server"))
	viper.BindPFlag("tools.integrationTests.runTests", integrationTestsCmd.Flags().Lookup("run"))
	viper.BindPFlag("tools.integrationTests.showSupported", integrationTestsCmd.Flags().Lookup("show-supported"))
	viper.BindEnv("tools.integrationTests.username", "LAL_USERNAME")
	viper.BindEnv("tools.integrationTests.password", "LAL_PASSWORD")
}

func supported() {
	const apiSupported = `
GET    /
GET    /version
POST   /alist
GET    /alist/:uuid
PUT    /alist/:uuid
DELETE /alist/:uuid
PUT    /alist/:uuid
`
	fmt.Println(apiSupported)
}
