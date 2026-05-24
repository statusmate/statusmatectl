package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/printer"
)

var ShowMaintenanceCmd = &cobra.Command{
	Use:     "show-maintenance <uuid>",
	Aliases: []string{"sm"},
	Short:   "Show detailed information about a maintenance",
	Args:    cobra.ExactArgs(1),
	RunE:    showMaintenanceCmdF,
}

func init() {
	ShowMaintenanceCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table, json")

	RootCmd.AddCommand(ShowMaintenanceCmd)
}

func showMaintenanceCmdF(command *cobra.Command, args []string) error {
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

	maintenance, err := client.GetMaintenanceByUUID(uuid)
	if err != nil {
		return err
	}

	return printer.PrintDetailMaintenance(os.Stdout, maintenance, format)
}
