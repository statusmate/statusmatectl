package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/printer"
)

var ListMaintenanceCmd = &cobra.Command{
	Use:   "list-maintenances",
	Short: "Ls command",
	RunE:  listMaintenancesCmdF,
}

func init() {
	ListMaintenanceCmd.Flags().BoolP("all", "a", false, "List active incidents")
	ListMaintenanceCmd.Flags().String("page", "", "Status page")
	ListMaintenanceCmd.Flags().String("format", printer.PrintTableFormatTable, "Format output")

	RootCmd.AddCommand(ListMaintenanceCmd)
}

func listMaintenancesCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	showAll, err := command.Flags().GetBool("all")
	if err != nil {
		return errors.Wrap(err, "all flag error")
	}

	statusPage, err := GetStatusPage(client, command)
	if err != nil {
		return errors.Wrap(err, "page flag error")
	}

	filters := api.PaginatedRequestFilter{
		"status":      api.MaintenanceActiveStatusList(),
		"status_page": statusPage.ID,
	}

	if showAll {
		delete(filters, "status")
	}

	payload := api.NewAllPaginatedRequest(filters)
	data, err := client.GetPaginatedMaintenance(payload)
	if err != nil {
		return err
	}

	config := printer.NewPrintTableConfig()
	config.PrintBlockTotal = false

	format, err := command.Flags().GetString("format")
	if err != nil {
		return errors.Wrap(err, "format flag error")
	}

	if err := printer.ValidatePrintTableFormat(format); err != nil {
		return err
	}

	config.Format = format

	err = printer.PrintMaintenances(os.Stdout, data, config)
	if err != nil {
		return err
	}

	return nil
}
