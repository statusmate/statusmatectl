package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/format"
)

var TouchIncidentCmd = &cobra.Command{
	Use:     "touch-incident [file]",
	Aliases: []string{"ti"},
	Short:   "Create an incident template file for later editing and submission",
	Args:    cobra.MaximumNArgs(1),
	RunE:    touchIncidentCmdF,
}

func init() {
	TouchIncidentCmd.Flags().StringP("page", "p", "", "Status page or default")
	RootCmd.AddCommand(TouchIncidentCmd)
}

func touchIncidentCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	statusPage, err := GetStatusPage(client, command)
	if err != nil {
		return err
	}

	filePath := "incident.inc"
	if len(args) > 0 {
		filePath = args[0]
	}

	payload := api.NewCreateIncidentPayload(statusPage)

	comps, err := client.GetPaginatedComponents(
		api.NewAllPaginatedRequest(api.PaginatedRequestFilter{"status_page": statusPage.ID}),
	)
	if err != nil {
		return err
	}

	data, err := format.Marshal(payload, &api.CreateIncidentPayloadFieldDescriptions)
	if err != nil {
		return err
	}

	data += api.BuildComponentsEditorFooter(comps.Results)

	if err := os.WriteFile(filePath, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Created: %s\n", filePath)
	return nil
}
