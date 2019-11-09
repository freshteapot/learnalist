package tools

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var stubSQLFileCmd = &cobra.Command{
	Use:   "stub-sql-file",
	Short: "Output the name of an sql file",
	Run: func(cmd *cobra.Command, args []string) {
		prefix := fmt.Sprintf("%s-XXX.sql", time.Now().UTC().Format("200601021504"))
		fmt.Println(prefix)
	},
}
