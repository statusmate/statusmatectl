package cmd

import (
	"github.com/spf13/cobra"
)

var UpdateIncidentCmd = &cobra.Command{
	Use:     "update-incident ",
	Aliases: []string{"i"},
	Short:   "Update incident",
	RunE:    updateIncidentCmdF,
}

func init() {
	UpdateIncidentCmd.Flags().String("uuid", "", "UUID incident")
	UpdateIncidentCmd.Flags().StringP("status", "s", "", "Update status")
	UpdateIncidentCmd.Flags().StringP("message", "m", "", "Message of the update")

	RootCmd.AddCommand(UpdateIncidentCmd)
}

func updateIncidentCmdF(command *cobra.Command, args []string) error {

	return nil
}
