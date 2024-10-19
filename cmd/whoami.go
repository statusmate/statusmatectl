package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:  "whoami",
	RunE: whoamiCmdF,
}

func init() {
	RootCmd.AddCommand(whoamiCmd)
}

func whoamiCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	fmt.Printf("Username: %s\n", client.Username)

	return nil
}
