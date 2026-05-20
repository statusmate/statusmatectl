package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
)

var CloseMaintenancesCmd = &cobra.Command{
	Use:     "close-maintenances",
	Aliases: []string{"complete-maintenances"},
	Short:   "Complete all active maintenances on a status page",
	RunE:    closeMaintenancesCmdF,
}

func init() {
	CloseMaintenancesCmd.Flags().StringP("page", "p", "", "Status page")
	CloseMaintenancesCmd.Flags().StringP("message", "m", "", "Completion message (optional)")
	CloseMaintenancesCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	CloseMaintenancesCmd.Flags().Bool("dry", false, "Dry run: show what would be completed without doing it")
	CloseMaintenancesCmd.Flags().Bool("notify", true, "Send notifications to subscribers")

	RootCmd.AddCommand(CloseMaintenancesCmd)
}

func closeMaintenancesCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	statusPage, err := GetStatusPage(client, command)
	if err != nil {
		return err
	}

	message, _ := command.Flags().GetString("message")
	yes, _ := command.Flags().GetBool("yes")
	dry, _ := command.Flags().GetBool("dry")
	notify, _ := command.Flags().GetBool("notify")

	maintenances, err := client.GetPaginatedMaintenance(
		api.NewAllPaginatedRequest(api.PaginatedRequestFilter{
			"status_page": statusPage.ID,
			"status":      api.MaintenanceActiveStatusList(),
		}),
	)
	if err != nil {
		return err
	}

	if maintenances.Count == 0 {
		fmt.Println("No active maintenances found.")
		return nil
	}

	fmt.Printf("Active maintenances (%d):\n", maintenances.Count)
	for _, m := range maintenances.Results {
		uuid := ""
		if m.UUID != nil {
			uuid = *m.UUID
		}
		fmt.Printf("  • [%s] %s (%s)\n", string(m.Status), m.Title, uuid)
	}

	if dry {
		fmt.Println("\nDry run: no maintenances were completed.")
		return nil
	}

	if !yes {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Complete all %d maintenance(s)?", maintenances.Count),
			IsConfirm: true,
		}
		if _, err := prompt.Run(); err != nil {
			fmt.Println("Canceled.")
			return nil
		}
	}

	fmt.Println()

	var failed []string
	for _, m := range maintenances.Results {
		uuid := ""
		if m.UUID != nil {
			uuid = *m.UUID
		}

		at := time.Now()
		if m.EndAt != nil && m.EndAt.Before(at) {
			at = *m.EndAt
		}

		update := &api.MaintenanceUpdate{
			Maintenance: m.ID,
			Status:      api.MaintenanceStatusCompleted,
			Description: message,
			Notify:      notify,
			At:          at,
			Components:  []api.AffectedComponent{},
		}

		if err := client.CreateMaintenanceUpdate(update); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s (%s): %v\n", m.Title, uuid, err)
			failed = append(failed, uuid)
			continue
		}

		fmt.Printf("  ✓ completed: %s (%s)\n", m.Title, uuid)
	}

	fmt.Println()
	if len(failed) > 0 {
		return fmt.Errorf("%d maintenance(s) failed to complete", len(failed))
	}

	fmt.Printf("Done. %d maintenance(s) completed.\n", maintenances.Count)
	return nil
}
