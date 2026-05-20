package cmd

import "github.com/spf13/cobra"

var LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "ls command",
}

func init() {
	RootCmd.AddCommand(LsCmd)
}
