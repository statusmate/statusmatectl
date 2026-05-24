package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/printer"
	"github.com/pkg/errors"
)

var ShowIncidentCmd = &cobra.Command{
	Use:     "show-incident [uuid]",
	Aliases: []string{"si"},
	Short:   "Show detailed information about an incident",
	Args:    cobra.MaximumNArgs(1),
	RunE:    showIncidentCmdF,
}

func init() {
	ShowIncidentCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table, json")
	ShowIncidentCmd.Flags().StringP("page", "p", "", "Status page")
	ShowIncidentCmd.Flags().BoolP("all", "a", false, "Include resolved incidents")

	RootCmd.AddCommand(ShowIncidentCmd)
}

func showIncidentCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	format, err := command.Flags().GetString("format")
	if err != nil {
		return err
	}
	if err := printer.ValidatePrintTableFormat(format); err != nil {
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
			"status":      api.IncidentActiveStatusList(),
			"status_page": statusPage.ID,
		}
		if showAll {
			delete(filters, "status")
		}

		incidents, err := client.GetPaginatedIncidents(api.NewAllPaginatedRequest(filters))
		if err != nil {
			return err
		}

		if incidents.Count == 0 {
			return nil
		}

		incident, err := pickIncident(incidents.Results)
		if err != nil {
			return err
		}
		if incident.UUID == nil {
			return nil
		}
		uuid = *incident.UUID
	}

	incident, err := client.GetIncidentByUUID(uuid)
	if err != nil {
		return err
	}

	return printer.PrintDetailIncident(os.Stdout, incident, format)
}
