package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
)

var CloseIncidentsCmd = &cobra.Command{
	Use:     "close-incidents",
	Aliases: []string{"resolve-incidents"},
	Short:   "Resolve all active incidents on a status page",
	RunE:    closeIncidentsCmdF,
}

func init() {
	CloseIncidentsCmd.Flags().StringP("page", "p", "", "Status page")
	CloseIncidentsCmd.Flags().StringP("message", "m", "", "Resolution message (optional)")
	CloseIncidentsCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	CloseIncidentsCmd.Flags().Bool("dry", false, "Dry run: show what would be resolved without doing it")
	CloseIncidentsCmd.Flags().Bool("notify", true, "Send notifications to subscribers")

	RootCmd.AddCommand(CloseIncidentsCmd)
}

func closeIncidentsCmdF(command *cobra.Command, args []string) error {
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

	incidents, err := client.GetPaginatedIncidents(
		api.NewAllPaginatedRequest(api.PaginatedRequestFilter{
			"status_page": statusPage.ID,
			"status":      api.IncidentActiveStatusList(),
		}),
	)
	if err != nil {
		return err
	}

	if incidents.Count == 0 {
		fmt.Println("No active incidents found.")
		return nil
	}

	fmt.Printf("Active incidents (%d):\n", incidents.Count)
	for _, inc := range incidents.Results {
		uuid := ""
		if inc.UUID != nil {
			uuid = *inc.UUID
		}
		fmt.Printf("  • [%s] %s (%s)\n", string(inc.Status), inc.Title, uuid)
	}

	if dry {
		fmt.Println("\nDry run: no incidents were resolved.")
		return nil
	}

	if !yes {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Resolve all %d incident(s)?", incidents.Count),
			IsConfirm: true,
		}
		if _, err := prompt.Run(); err != nil {
			fmt.Println("Canceled.")
			return nil
		}
	}

	fmt.Println()

	var failed []string
	for _, inc := range incidents.Results {
		uuid := ""
		if inc.UUID != nil {
			uuid = *inc.UUID
		}

		update := &api.IncidentUpdate{
			Incident:    inc.ID,
			Status:      api.IncidentStatusResolved,
			Description: message,
			Notify:      notify,
			At:          time.Now(),
			Components:  []api.AffectedComponent{},
		}

		if err := client.CreateIncidentUpdate(update); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s (%s): %v\n", inc.Title, uuid, err)
			failed = append(failed, uuid)
			continue
		}

		fmt.Printf("  ✓ resolved: %s (%s)\n", inc.Title, uuid)
	}

	fmt.Println()
	if len(failed) > 0 {
		return fmt.Errorf("%d incident(s) failed to resolve", len(failed))
	}

	fmt.Printf("Done. %d incident(s) resolved.\n", incidents.Count)
	return nil
}
