package challenges

import (
	"io"
	"log"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "challenges",
	Short: "Challenge commands",
}

func init() {
	RootCmd.AddCommand(syncCMD)
}

func logCloser(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("close error: %s", err)
	}
}
