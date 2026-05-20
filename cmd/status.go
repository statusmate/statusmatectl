package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/printer"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show component status tree with incident/maintenance reasons",
	RunE:  statusCmdF,
}

func init() {
	StatusCmd.Flags().StringP("page", "p", "", "Status page")
	RootCmd.AddCommand(StatusCmd)
}

func statusCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	statusPage, err := GetStatusPage(client, command)
	if err != nil {
		return err
	}

	components, err := client.GetPaginatedComponents(
		api.NewAllPaginatedRequest(api.PaginatedRequestFilter{"status_page": statusPage.ID}),
	)
	if err != nil {
		return err
	}

	incidents, err := client.GetPaginatedIncidents(
		api.NewAllPaginatedRequest(api.PaginatedRequestFilter{
			"status_page": statusPage.ID,
			"status":      api.IncidentActiveStatusList(),
		}),
	)
	if err != nil {
		return err
	}

	maintenances, err := client.GetPaginatedMaintenance(
		api.NewAllPaginatedRequest(api.PaginatedRequestFilter{
			"status_page": statusPage.ID,
			"status":      api.MaintenanceActiveStatusList(),
		}),
	)
	if err != nil {
		return err
	}

	// componentID → []reason (deduplicated per incident/maintenance)
	reasons := make(map[int][]string)
	seen := make(map[[2]string]bool) // [componentID_str, title] → added

	addReason := func(componentID int, label string) {
		key := [2]string{fmt.Sprintf("%d", componentID), label}
		if !seen[key] {
			seen[key] = true
			reasons[componentID] = append(reasons[componentID], label)
		}
	}

	// компоненты живут внутри updates, а не на верхнем уровне инцидента
	for _, inc := range incidents.Results {
		for _, update := range inc.Updates {
			for _, ac := range update.Components {
				addReason(ac.Component, "🚨 "+inc.Title)
			}
		}
	}

	for _, m := range maintenances.Results {
		for _, update := range m.Updates {
			for _, ac := range update.Components {
				addReason(ac.Component, "🔧 "+m.Title)
			}
		}
	}

	return printer.PrintComponentStatusTree(os.Stdout, components, reasons)
}
