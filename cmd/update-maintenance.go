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

// maintenance update [uuid]
var maintenanceUpdateSubCmd = &cobra.Command{
	Use:   "update [uuid]",
	Short: "Post an update to a maintenance",
	Args:  cobra.MaximumNArgs(1),
	RunE:  maintenanceUpdateSubCmdF,
}

func init() {
	maintenanceUpdateSubCmd.Flags().StringP("status", "s", "", "Update status")
	maintenanceUpdateSubCmd.Flags().StringP("desc", "d", "", "Update message/description")
	maintenanceUpdateSubCmd.Flags().StringArrayP("components", "c", []string{}, "Affected components in 'impact component' format")
	maintenanceUpdateSubCmd.Flags().Bool("notify", true, "Send notifications")
	maintenanceUpdateSubCmd.Flags().BoolP("interactive", "i", false, "Interactive editing mode (editor)")
	maintenanceUpdateSubCmd.Flags().BoolP("pick-components", "C", false, "Interactively pick components")
	maintenanceUpdateSubCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	maintenanceUpdateSubCmd.Flags().BoolP("next", "n", false, "Advance status to next step")
	maintenanceUpdateSubCmd.Flags().StringP("page", "p", "", "Status page (used for maintenance picker)")
	maintenanceUpdateSubCmd.Flags().BoolP("all", "a", false, "Include completed maintenances in picker")

	MaintenanceCmd.AddCommand(maintenanceUpdateSubCmd)
}

func maintenanceUpdateSubCmdF(command *cobra.Command, args []string) error {
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

		picked, err := pickMaintenance(maintenances.Results)
		if err != nil {
			return err
		}
		if picked.UUID == nil {
			return nil
		}
		uuid = *picked.UUID
	}

	maintenance, err := client.GetMaintenanceByUUID(uuid)
	if err != nil {
		return err
	}

	comps, err := client.GetPaginatedComponents(
		api.NewAllPaginatedRequest(api.PaginatedRequestFilter{"status_page": maintenance.StatusPage}),
	)
	if err != nil {
		return err
	}
	availableComponents := comps.Results

	var sourceComponents []api.AffectedComponent
	if maintenance.ID != nil {
		latestUpdate, err := client.GetLatestMaintenanceUpdate(*maintenance.ID)
		if err != nil {
			return err
		}
		if latestUpdate != nil {
			sourceComponents = latestUpdate.Components
		}
	}
	if sourceComponents == nil {
		sourceComponents = maintenance.Components
	}

	payload := &api.CreateMaintenanceUpdatePayload{
		Status:     string(maintenance.Status),
		Components: affectedComponentsToStrings(sourceComponents, availableComponents),
		Notify:     true,
	}

	if next, _ := command.Flags().GetBool("next"); next {
		nextStatus, err := api.NextMaintenanceStatus(api.MaintenanceStatusType(payload.Status))
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

	if interactive {
		data, err := format.Marshal(payload, &api.CreateMaintenanceUpdatePayloadFieldDescriptions)
		if err != nil {
			return err
		}
		data += api.BuildComponentsEditorFooter(availableComponents)

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

	update := &api.MaintenanceUpdate{
		Maintenance: maintenance.ID,
		Status:      api.MaintenanceStatusType(payload.Status),
		Description: payload.Description,
		Notify:      payload.Notify,
		At:          time.Now(),
		Components:  affectedComps,
	}

	if err := client.CreateMaintenanceUpdate(update); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "maintenance=%s\nstatus=%s\n", uuid, payload.Status)
	return nil
}
