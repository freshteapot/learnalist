package notifications

import (
	"io"
	"log"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "notifications",
	Short: "notification commands",
}

func init() {
	RootCmd.AddCommand(pushNotificationsCMD)
}

func logCloser(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("close error: %s", err)
	}
}
