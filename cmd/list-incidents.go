package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/printer"
)

var ListIncidentsCmd = &cobra.Command{
	Use:   "list-incidents",
	Short: "Ls command",
	RunE:  listIncidentsCmdF,
}

func init() {
	ListIncidentsCmd.Flags().BoolP("all", "a", false, "List active incidents")
	ListIncidentsCmd.Flags().String("page", "", "Status page")
	ListIncidentsCmd.Flags().String("format", printer.PrintTableFormatTable, "Format output")

	RootCmd.AddCommand(ListIncidentsCmd)
}

func listIncidentsCmdF(command *cobra.Command, args []string) error {
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
		"status":      api.IncidentActiveStatusList(),
		"status_page": statusPage.ID,
	}

	if showAll {
		delete(filters, "status")
	}

	payload := api.NewAllPaginatedRequest(filters)
	data, err := client.GetPaginatedIncidents(payload)
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

	err = printer.PrintIncidents(os.Stdout, data, config)
	if err != nil {
		return err
	}

	return nil
}
