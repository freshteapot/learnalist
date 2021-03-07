package user

import (
	"fmt"

	"github.com/spf13/cobra"
)

var scratchCMD = &cobra.Command{
	Use:   "scratch",
	Short: "Temp code",
	Long: `

	go run --tags="json1"  main.go --config=../config/dev.config.yaml tools user scratch
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO")
	},
}
