package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/format"
)

var TouchMaintenanceCmd = &cobra.Command{
	Use:     "touch-maintenance [file]",
	Aliases: []string{"tm"},
	Short:   "Create a maintenance template file for later editing and submission",
	Args:    cobra.MaximumNArgs(1),
	RunE:    touchMaintenanceCmdF,
}

func init() {
	TouchMaintenanceCmd.Flags().StringP("page", "p", "", "Status page or default")
	RootCmd.AddCommand(TouchMaintenanceCmd)
}

func touchMaintenanceCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	statusPage, err := GetStatusPage(client, command)
	if err != nil {
		return err
	}

	filePath := "maintenance.mnc"
	if len(args) > 0 {
		filePath = args[0]
	}

	payload := api.NewCreateMaintenancePayload(statusPage)

	comps, err := client.GetPaginatedComponents(
		api.NewAllPaginatedRequest(api.PaginatedRequestFilter{"status_page": statusPage.ID}),
	)
	if err != nil {
		return err
	}

	data, err := format.Marshal(payload, &api.CreateMaintenancePayloadFieldDescriptions)
	if err != nil {
		return err
	}

	data += buildComponentsFooter(comps.Results)

	if err := os.WriteFile(filePath, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Created: %s\n", filePath)
	return nil
}
