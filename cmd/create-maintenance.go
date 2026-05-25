package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/editor"
	"github.com/statusmate/statusmatectl/pkg/format"
	"github.com/statusmate/statusmatectl/pkg/printer"
)

var CreateMaintenanceCmd = &cobra.Command{
	Use:     "create-maintenance",
	Aliases: []string{"cm"},
	Short:   "Schedule maintenance",
	RunE:    createMaintenanceCmdF,
}

func init() {
	CreateMaintenanceCmd.Flags().StringP("page", "p", "", "Status page or default")
	CreateMaintenanceCmd.Flags().StringP("title", "n", "", "Title of the maintenance")
	CreateMaintenanceCmd.Flags().StringP("desc", "d", "", "Description")
	CreateMaintenanceCmd.Flags().String("start-at", "", "Start time (RFC3339)")
	CreateMaintenanceCmd.Flags().String("end-at", "", "End time (RFC3339, optional)")
	CreateMaintenanceCmd.Flags().StringArrayP("components", "c", []string{}, "Affected components in 'impact component' format")
	CreateMaintenanceCmd.Flags().Bool("notify", true, "Send notifications to subscribers")
	CreateMaintenanceCmd.Flags().Bool("auto-start", false, "Auto-start at start_at time")
	CreateMaintenanceCmd.Flags().Bool("auto-end", false, "Auto-end at end_at time")
	CreateMaintenanceCmd.Flags().Bool("affect-uptime", true, "Count against uptime")
	CreateMaintenanceCmd.Flags().BoolP("interactive", "i", false, "Interactive editing mode")
	CreateMaintenanceCmd.Flags().BoolP("pick-components", "C", false, "Interactively pick components")
	CreateMaintenanceCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	CreateMaintenanceCmd.Flags().Bool("dry", false, "Dry run: validate and show without creating")
	CreateMaintenanceCmd.Flags().StringP("file", "f", "", "Path to maintenance template file (created with touch-maintenance)")

	RootCmd.AddCommand(CreateMaintenanceCmd)
}

func createMaintenanceCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	statusPage, err := GetStatusPage(client, command)
	if err != nil {
		return err
	}

	title, _ := command.Flags().GetString("title")
	description, _ := command.Flags().GetString("desc")
	startAt, _ := command.Flags().GetString("start-at")
	endAt, _ := command.Flags().GetString("end-at")
	components, _ := command.Flags().GetStringArray("components")
	notify, _ := command.Flags().GetBool("notify")
	autoStart, _ := command.Flags().GetBool("auto-start")
	autoEnd, _ := command.Flags().GetBool("auto-end")
	affectUptime, _ := command.Flags().GetBool("affect-uptime")
	interactive, _ := command.Flags().GetBool("interactive")
	pickComponents, _ := command.Flags().GetBool("pick-components")
	dry, _ := command.Flags().GetBool("dry")
	yes, _ := command.Flags().GetBool("yes")
	filePath, _ := command.Flags().GetString("file")

	payload := api.NewCreateMaintenancePayload(statusPage)

	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return errors.Wrapf(err, "failed to read file %q", filePath)
		}
		if err := format.Unmarshal(string(data), payload); err != nil {
			return errors.Wrapf(err, "failed to parse file %q", filePath)
		}
	} else {
		payload.Title = title
		payload.Description = description
		payload.Components = components
		payload.Notify = notify
		payload.AutoStart = autoStart
		payload.AutoEnd = autoEnd
		payload.AffectUptime = affectUptime
		if startAt != "" {
			payload.StartAt = startAt
		}
		if endAt != "" {
			payload.EndAt = endAt
		}
	}

	var availableComponents []api.Component
	if interactive || pickComponents {
		comps, err := client.GetPaginatedComponents(
			api.NewAllPaginatedRequest(api.PaginatedRequestFilter{"status_page": statusPage.ID}),
		)
		if err != nil {
			return err
		}
		availableComponents = comps.Results
	}

	if pickComponents {
		picked, err := selectComponentsInteractive(availableComponents)
		if err != nil {
			return err
		}
		payload.Components = append(payload.Components, picked...)
	}

	if interactive {
		data, err := format.Marshal(payload, &api.CreateMaintenancePayloadFieldDescriptions)
		if err != nil {
			return err
		}
		data += buildComponentsFooter(availableComponents)

		for {
			output, err := editor.CaptureInputFromEditor([]byte(data))
			if err != nil {
				return err
			}

			err = format.Unmarshal(string(output), payload)
			if err != nil {
				return err
			}

			validationErrs := validateCreateMaintenancePayload(payload)
			if len(validationErrs) == 0 {
				var buf bytes.Buffer
				printer.PrintSummaryCreateMaintenancePayload(&buf, payload)
				fmt.Printf("Maintenance\n%s", buf.String())
				break
			}

			fmt.Println("Ошибки валидации:")
			for _, e := range validationErrs {
				fmt.Printf("  • %s\n", e)
			}

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

	if dry {
		fmt.Println("Dry-run mode enabled. Maintenance not created.")
		return nil
	}

	if !yes {
		prompt := promptui.Prompt{
			Label:     "Schedule maintenance?",
			IsConfirm: true,
		}
		_, err := prompt.Run()
		if err != nil {
			fmt.Println("Canceled.")
			return nil
		}
	}

	maintenance, err := client.CreateMaintenance(payload)
	if err != nil {
		return err
	}

	printer.PrintSummaryMaintenance(os.Stdout, maintenance)
	return nil
}

func validateCreateMaintenancePayload(p *api.CreateMaintenancePayload) []string {
	var errs []string
	if strings.TrimSpace(p.Title) == "" {
		errs = append(errs, "title: обязательное поле")
	}
	if strings.TrimSpace(p.Description) == "" {
		errs = append(errs, "description: обязательное поле")
	}
	if len(p.Components) == 0 {
		errs = append(errs, "components: укажите хотя бы один компонент")
	}
	if strings.TrimSpace(p.StartAt) == "" {
		errs = append(errs, "start_at: обязательное поле")
	}
	return errs
}
