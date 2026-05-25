package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
)

var OpenMaintenanceCmd = &cobra.Command{
	Use:     "open-maintenance [uuid]",
	Aliases: []string{"om"},
	Short:   "Open a maintenance in the browser",
	Args:    cobra.MaximumNArgs(1),
	RunE:    openMaintenanceCmdF,
}

var ShortOpenMaintenanceCmd = &cobra.Command{
	Use:   "om [uuid]",
	Short: "Open a maintenance in the browser",
	Args:  cobra.MaximumNArgs(1),
	RunE:  openMaintenanceCmdF,
}

func init() {
	for _, cmd := range []*cobra.Command{OpenMaintenanceCmd, ShortOpenMaintenanceCmd} {
		cmd.Flags().StringP("page", "p", "", "Status page")
		cmd.Flags().BoolP("all", "a", false, "Include completed maintenances")
	}
	RootCmd.AddCommand(OpenMaintenanceCmd)
	LsCmd.AddCommand(ShortOpenMaintenanceCmd)
}

func openMaintenanceCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	statusPage, err := GetStatusPage(client, command)
	if err != nil {
		return err
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
		fmt.Println("No maintenances found.")
		return nil
	}

	var m *api.Maintenance

	if len(args) == 1 {
		uuid := args[0]
		for i := range maintenances.Results {
			if maintenances.Results[i].UUID != nil && *maintenances.Results[i].UUID == uuid {
				m = &maintenances.Results[i]
				break
			}
		}
		if m == nil {
			return fmt.Errorf("maintenance not found: %s", uuid)
		}
	} else {
		m, err = pickMaintenance(maintenances.Results)
		if err != nil {
			return err
		}
	}

	url := m.AbsoluteURL
	if url == "" {
		return fmt.Errorf("maintenance has no URL")
	}

	fmt.Println(url)

	var openCmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		openCmd = exec.Command("open", url)
	} else {
		openCmd = exec.Command("xdg-open", url)
	}
	return openCmd.Start()
}

func pickMaintenance(maintenances []api.Maintenance) (*api.Maintenance, error) {
	items := make([]string, len(maintenances))
	for i, m := range maintenances {
		uuid := ""
		if m.UUID != nil {
			uuid = *m.UUID
		}
		items[i] = fmt.Sprintf("[%s] %s (%s)", m.Status, m.Title, uuid)
	}

	prompt := promptui.Select{
		Label: "Select maintenance",
		Items: items,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &maintenances[idx], nil
}
