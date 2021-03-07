package alist

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "list",
	Short: "Management of list via the cli",
}

func init() {
	RootCmd.AddCommand(publicAccessCMD)
}
