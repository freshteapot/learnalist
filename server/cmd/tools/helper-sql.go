package tools

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var stubSQLFileCmd = &cobra.Command{
	Use:   "stub-sql-file",
	Short: "Output the name of an sql file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		path := fmt.Sprintf("db/%s-%s.sql", time.Now().UTC().Format("200601021504"), name)
		fmt.Println(path)
	},
}
