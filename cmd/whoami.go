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

	user, err := client.GetMe()
	if err != nil {
		return err
	}

	fmt.Printf("Server:   %s\n", client.BaseURL)
	fmt.Printf("Username: %s\n", user.Username)
	fmt.Printf("Email:    %s\n", user.Email)

	return nil
}
