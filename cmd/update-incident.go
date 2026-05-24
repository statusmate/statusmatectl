package cmd

import (
	"github.com/spf13/cobra"
)

var UpdateIncidentCmd = &cobra.Command{
	Use:     "update-incident ",
	Short:   "Update incident",
	RunE:    updateIncidentCmdF,
}

func init() {
	RootCmd.AddCommand(UpdateIncidentCmd)
}

func updateIncidentCmdF(command *cobra.Command, args []string) error {

	return nil
}
