package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/printer"
)

var ShowIncidentCmd = &cobra.Command{
	Use:     "show-incident <uuid>",
	Aliases: []string{"si"},
	Short:   "Show detailed information about an incident",
	Args:    cobra.ExactArgs(1),
	RunE:    showIncidentCmdF,
}

func init() {
	ShowIncidentCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table, json")

	RootCmd.AddCommand(ShowIncidentCmd)
}

func showIncidentCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	uuid := args[0]

	format, err := command.Flags().GetString("format")
	if err != nil {
		return err
	}
	if err := printer.ValidatePrintTableFormat(format); err != nil {
		return err
	}

	incident, err := client.GetIncidentByUUID(uuid)
	if err != nil {
		return err
	}

	return printer.PrintDetailIncident(os.Stdout, incident, format)
}
