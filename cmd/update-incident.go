package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/editor"
	"github.com/statusmate/statusmatectl/pkg/format"
)

// incident update [uuid]
var incidentUpdateSubCmd = &cobra.Command{
	Use:   "update [uuid]",
	Short: "Update incident metadata",
	Args:  cobra.MaximumNArgs(1),
	RunE:  incidentUpdateSubCmdF,
}

func init() {
	incidentUpdateSubCmd.Flags().StringP("status", "s", "", "Update status")
	incidentUpdateSubCmd.Flags().StringP("desc", "d", "", "Update message/description")
	incidentUpdateSubCmd.Flags().StringArrayP("components", "c", []string{}, "Affected components in 'impact component' format")
	incidentUpdateSubCmd.Flags().Bool("notify", true, "Send notifications")
	incidentUpdateSubCmd.Flags().BoolP("interactive", "i", false, "Interactive editing mode (editor)")
	incidentUpdateSubCmd.Flags().BoolP("pick-components", "C", false, "Interactively pick components")
	incidentUpdateSubCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	incidentUpdateSubCmd.Flags().BoolP("next", "n", false, "Advance status to next step")
	incidentUpdateSubCmd.Flags().StringP("page", "p", "", "Status page (used for incident picker)")
	incidentUpdateSubCmd.Flags().BoolP("all", "a", false, "Include resolved incidents in picker")

	IncidentCmd.AddCommand(incidentUpdateSubCmd)
}

func incidentUpdateSubCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	var uuid string
	if len(args) == 1 {
		uuid = args[0]
	} else {
		statusPage, err := GetStatusPage(client, command)
		if err != nil {
			return fmt.Errorf("page flag error: %w", err)
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

		picked, err := pickIncident(incidents.Results)
		if err != nil {
			return err
		}
		if picked.UUID == nil {
			return nil
		}
		uuid = *picked.UUID
	}

	incident, err := client.GetIncidentByUUID(uuid)
	if err != nil {
		return err
	}

	comps, err := client.GetPaginatedComponents(
		api.NewAllPaginatedRequest(api.PaginatedRequestFilter{"status_page": incident.StatusPage}),
	)
	if err != nil {
		return err
	}
	availableComponents := comps.Results

	var sourceComponents []api.AffectedComponent
	if incident.ID != nil {
		latestUpdate, err := client.GetLatestIncidentUpdate(*incident.ID)
		if err != nil {
			return err
		}
		if latestUpdate != nil {
			sourceComponents = latestUpdate.Components
		}
	}
	if sourceComponents == nil {
		sourceComponents = incident.Components
	}

	payload := &api.CreateIncidentUpdatePayload{
		Status:     string(incident.Status),
		Components: affectedComponentsToStrings(sourceComponents, availableComponents),
		Notify:     true,
	}

	if next, _ := command.Flags().GetBool("next"); next {
		nextStatus, err := api.NextIncidentStatus(api.IncidentStatusType(payload.Status))
		if err != nil {
			return err
		}
		payload.Status = string(nextStatus)
	}
	if command.Flags().Changed("status") {
		payload.Status, _ = command.Flags().GetString("status")
	}
	if command.Flags().Changed("desc") {
		payload.Description, _ = command.Flags().GetString("desc")
	}
	if command.Flags().Changed("components") {
		payload.Components, _ = command.Flags().GetStringArray("components")
	}
	if command.Flags().Changed("notify") {
		payload.Notify, _ = command.Flags().GetBool("notify")
	}

	interactive, _ := command.Flags().GetBool("interactive")
	pickComponents, _ := command.Flags().GetBool("pick-components")

	if pickComponents {
		picked, err := selectComponentsInteractive(availableComponents)
		if err != nil {
			return err
		}
		payload.Components = append(payload.Components, picked...)
	}

	if api.IncidentStatusType(payload.Status) == api.IncidentStatusResolved &&
		!command.Flags().Changed("components") && !pickComponents {
		payload.Components = resolveAllComponents(sourceComponents, availableComponents)
	}

	if interactive {
		data, err := format.Marshal(payload, &api.CreateIncidentUpdatePayloadFieldDescriptions)
		if err != nil {
			return err
		}
		data += buildComponentsFooter(availableComponents)

		for {
			output, err := editor.CaptureInputFromEditor([]byte(data))
			if err != nil {
				return err
			}

			if err := format.Unmarshal(string(output), payload); err != nil {
				return err
			}

			if strings.TrimSpace(payload.Description) != "" {
				break
			}

			fmt.Println("Ошибки валидации:")
			fmt.Println("  • description: обязательное поле")

			sel := promptui.Select{
				Label: "Продолжить?",
				Items: []string{"Редактировать снова", "Отменить"},
			}
			idx, _, err := sel.Run()
			if err != nil || idx == 1 {
				fmt.Println("Отменено.")
				return nil
			}
			data = string(output)
		}
	}

	yes, _ := command.Flags().GetBool("yes")
	if !yes {
		prompt := promptui.Prompt{
			Label:     "Post update?",
			IsConfirm: true,
		}
		if _, err := prompt.Run(); err != nil {
			fmt.Println("Canceled.")
			return nil
		}
	}

	var affectedComps []api.AffectedComponent
	if len(payload.Components) > 0 {
		affectedComps, err = api.BuildAffectedComponents(payload.Components, availableComponents)
		if err != nil {
			return err
		}
	}

	update := &api.IncidentUpdate{
		Incident:    incident.ID,
		Status:      api.IncidentStatusType(payload.Status),
		Description: payload.Description,
		Notify:      payload.Notify,
		At:          time.Now(),
		Components:  affectedComps,
	}

	if err := client.CreateIncidentUpdate(update); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "incident=%s\nstatus=%s\n", uuid, payload.Status)
	return nil
}

func resolveAllComponents(comps []api.AffectedComponent, available []api.Component) []string {
	result := make([]string, 0, len(comps))
	for _, ac := range comps {
		for _, c := range available {
			if c.ID != nil && *c.ID == ac.Component {
				result = append(result, "operational "+c.Name)
				break
			}
		}
	}
	return result
}

func affectedComponentsToStrings(comps []api.AffectedComponent, available []api.Component) []string {
	result := make([]string, 0, len(comps))
	for _, ac := range comps {
		var name string
		for _, c := range available {
			if c.ID != nil && *c.ID == ac.Component {
				name = c.Name
				break
			}
		}
		if name == "" {
			continue
		}
		result = append(result, fmt.Sprintf("%s %s", ac.Impact, name))
	}
	return result
}
