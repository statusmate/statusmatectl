package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/printer"
)

var ShowMaintenanceCmd = &cobra.Command{
	Use:     "show-maintenance [uuid]",
	Aliases: []string{"sm"},
	Short:   "Show detailed information about a maintenance",
	Args:    cobra.MaximumNArgs(1),
	RunE:    showMaintenanceCmdF,
}

func init() {
	ShowMaintenanceCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table, json")
	ShowMaintenanceCmd.Flags().StringP("page", "p", "", "Status page")
	ShowMaintenanceCmd.Flags().BoolP("all", "a", false, "Include completed maintenances")

	RootCmd.AddCommand(ShowMaintenanceCmd)
}

func showMaintenanceCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	outputFormat, err := command.Flags().GetString("format")
	if err != nil {
		return err
	}
	if err := printer.ValidatePrintTableFormat(outputFormat); err != nil {
		return err
	}

	var uuid string
	if len(args) == 1 {
		uuid = args[0]
	} else {
		statusPage, err := GetStatusPage(client, command)
		if err != nil {
			return errors.Wrap(err, "page flag error")
		}

		showAll, _ := command.Flags().GetBool("all")
		filters := api.PaginatedRequestFilter{
			"status":      api.MaintenanceActiveStatusList(),
			"status_page": statusPage.ID,
		}
		if showAll {
			delete(filters, "status")
		}

		maintenances, err := client.GetPaginatedMaintenance(api.NewAllPaginatedRequest(filters))
		if err != nil {
			return err
		}

		if maintenances.Count == 0 {
			return nil
		}

		m, err := pickMaintenance(maintenances.Results)
		if err != nil {
			return err
		}
		if m.UUID == nil {
			return nil
		}
		uuid = *m.UUID
	}

	maintenance, err := client.GetMaintenanceByUUID(uuid)
	if err != nil {
		return err
	}

	return printer.PrintDetailMaintenance(os.Stdout, maintenance, outputFormat)
}
