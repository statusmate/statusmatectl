package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
)

var OpenIncidentCmd = &cobra.Command{
	Use:     "open-incident [uuid]",
	Aliases: []string{"oi"},
	Short:   "Open an incident in the browser",
	Args:    cobra.MaximumNArgs(1),
	RunE:    openIncidentCmdF,
}

var ShortOpenIncidentCmd = &cobra.Command{
	Use:   "oi [uuid]",
	Short: "Open an incident in the browser",
	Args:  cobra.MaximumNArgs(1),
	RunE:  openIncidentCmdF,
}

func init() {
	for _, cmd := range []*cobra.Command{OpenIncidentCmd, ShortOpenIncidentCmd} {
		cmd.Flags().StringP("page", "p", "", "Status page")
		cmd.Flags().BoolP("all", "a", false, "Include resolved incidents")
	}
	RootCmd.AddCommand(OpenIncidentCmd)
	LsCmd.AddCommand(ShortOpenIncidentCmd)
}

func openIncidentCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

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
		fmt.Println("No incidents found.")
		return nil
	}

	var incident *api.Incident

	if len(args) == 1 {
		uuid := args[0]
		for i := range incidents.Results {
			if incidents.Results[i].UUID != nil && *incidents.Results[i].UUID == uuid {
				incident = &incidents.Results[i]
				break
			}
		}
		if incident == nil {
			return fmt.Errorf("incident not found: %s", uuid)
		}
	} else {
		incident, err = pickIncident(incidents.Results)
		if err != nil {
			return err
		}
	}

	if incident.AbsoluteURL == nil || *incident.AbsoluteURL == "" {
		return fmt.Errorf("incident has no URL")
	}

	url := *incident.AbsoluteURL
	fmt.Println(url)

	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", url)
	} else {
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func pickIncident(incidents []api.Incident) (*api.Incident, error) {
	items := make([]string, len(incidents))
	for i, inc := range incidents {
		uuid := ""
		if inc.UUID != nil {
			uuid = *inc.UUID
		}
		items[i] = fmt.Sprintf("[%s] %s (%s)", inc.Status, inc.Title, uuid)
	}

	prompt := promptui.Select{
		Label: "Select incident",
		Items: items,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &incidents[idx], nil
}
