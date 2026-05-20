package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/editor"
	"github.com/statusmate/statusmatectl/pkg/format"
	"github.com/statusmate/statusmatectl/pkg/printer"
)

var CreateIncidentCmd = &cobra.Command{
	Use:     "create-incident",
	Aliases: []string{"ci"},
	Short:   "Create incident",
	RunE:    createIncidentCmdF,
}

/**
 *
 * echo ./incident.json | st4 ci
 */
var CreateIncidentStdinCmd = &cobra.Command{
	Use:   "stdin",
	Short: "Create incident from stdin",
	RunE:  createIncidentCmdF,
}

/*
	*

*
Output:

Incident "Database Outage" created successfully!

Summary:
uuid=12345
name=Database Outage
description=All DB servers are down
components=DB, API
status=Investigating
created_at=2024-10-22 12:34

Такой ответ будет удобно читать через awk
| awk -F= '/uuid/ {print $2}'
*/
func init() {
	//st4 create-incident --page <page_id>
	CreateIncidentCmd.Flags().StringP("page", "p", "", "Status page or default")
	//st4 create-incident --status investigation
	CreateIncidentCmd.Flags().StringP("status", "s", string(api.IncidentStatusInvestigation), "Status")
	//st4 create-incident --title "Проблема"
	CreateIncidentCmd.Flags().StringP("title", "n", "", "Title of the incident")
	//st4 create-incident --desc "Проблема"
	CreateIncidentCmd.Flags().StringP("desc", "d", "", "Description of the incident")
	//st4 create-incident  --components="o cdn" --components="p web"
	CreateIncidentCmd.Flags().StringArrayP("components", "c", []string{}, "Specify components with impact in 'impact component' format")
	//st4 create-incident  --private="Private message"
	CreateIncidentCmd.Flags().String("private", "", "Affected components, e.g. op cloud/lkk")
	//st4 create-incident  --showOnTop
	CreateIncidentCmd.Flags().Bool("showOnTop", false, "Affected components, e.g. op cloud/lkk")
	//st4 create-incident  --notify
	CreateIncidentCmd.Flags().Bool("notify", true, "Send notify")
	//st4 create-incident  --interactive
	CreateIncidentCmd.Flags().BoolP("interactive", "i", false, "Enable interactive editing mode")
	//st4 create-incident  --pick-components
	CreateIncidentCmd.Flags().BoolP("pick-components", "C", false, "Interactively pick components and their impact")
	//st4 create-incident  -y
	CreateIncidentCmd.Flags().BoolP("yes", "y", false, "yes for prompt")
	//st4 create-incident  --dry
	CreateIncidentCmd.Flags().Bool("dry", false, "Dry run, check data and open editor from $EDITOR")
	//st4 create-incident -f incident.inc
	CreateIncidentCmd.Flags().StringP("file", "f", "", "Path to incident template file (created with touch-incident)")

	RootCmd.AddCommand(CreateIncidentCmd)
}

func createIncidentCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	statusPage, err := GetStatusPage(client, command)
	if err != nil {
		return err
	}

	title, _ := command.Flags().GetString("title")
	status, _ := command.Flags().GetString("status")
	description, _ := command.Flags().GetString("desc")
	components, _ := command.Flags().GetStringArray("components")
	interactive, _ := command.Flags().GetBool("interactive")
	pickComponents, _ := command.Flags().GetBool("pick-components")
	dry, _ := command.Flags().GetBool("dry")
	yes, _ := command.Flags().GetBool("yes")
	filePath, _ := command.Flags().GetString("file")

	newIncident := api.NewCreateIncidentPayload(statusPage)

	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", filePath, err)
		}
		if err := format.Unmarshal(string(data), newIncident); err != nil {
			return fmt.Errorf("failed to parse file %q: %w", filePath, err)
		}
		newIncident.StartAt = time.Now()
	} else {
		newIncident.Title = title
		newIncident.Components = components
		newIncident.Description = description
		newIncident.Status = status
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
		newIncident.Components = append(newIncident.Components, picked...)
	}

	if interactive {
		data, err := format.Marshal(newIncident, &api.CreateIncidentPayloadFieldDescriptions)
		if err != nil {
			return err
		}
		data += buildComponentsFooter(availableComponents)

		for {
			output, err := editor.CaptureInputFromEditor([]byte(data))
			if err != nil {
				return err
			}

			err = format.Unmarshal(string(output), newIncident)
			if err != nil {
				return err
			}

			validationErrs := validateCreateIncidentPayload(newIncident)
			if len(validationErrs) == 0 {
				var buf bytes.Buffer
				printer.PrintSummaryCreateIncidentPayload(&buf, newIncident)
				fmt.Printf("Incident\n%s", buf.String())
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

			// сохраняем отредактированный вывод как основу для следующего прохода
			data = string(output)
		}
	}

	if dry {
		fmt.Println("Dry-run mode enabled. Incident not created.")
		return nil
	}

	if !yes {
		prompt := promptui.Prompt{
			Label:     "Create incident?",
			IsConfirm: true,
		}

		_, err := prompt.Run()
		if err != nil {
			fmt.Println("Canceled.")
			return nil
		}
	}

	incident, err := client.CreateIncident(newIncident)
	if err != nil {
		return err
	}

	printer.PrintSummaryIncident(os.Stdout, incident)

	return nil
}

var impactChoices = []struct {
	label string
	key   string
}{
	{"p  — partial outage", "p"},
	{"m  — major outage", "m"},
	{"d  — degraded performance", "d"},
	{"u  — under maintenance", "u"},
	{"o  — operational", "o"},
}

type componentEntry struct {
	component api.Component
	display   string // name with indentation reflecting hierarchy
}

func flattenComponentTree(components []api.Component) []componentEntry {
	children := make(map[int][]api.Component)
	var roots []api.Component

	for _, c := range components {
		if c.Parent == nil {
			roots = append(roots, c)
		} else {
			children[*c.Parent] = append(children[*c.Parent], c)
		}
	}

	var result []componentEntry
	var flatten func(comps []api.Component, indent string)
	flatten = func(comps []api.Component, indent string) {
		for _, c := range comps {
			result = append(result, componentEntry{component: c, display: indent + c.Name})
			if c.ID != nil {
				if kids, ok := children[*c.ID]; ok {
					flatten(kids, indent+"  ")
				}
			}
		}
	}
	flatten(roots, "")

	return result
}

func buildComponentsFooter(components []api.Component) string {
	if len(components) == 0 {
		return ""
	}

	children := make(map[int][]api.Component)
	var roots []api.Component
	for _, c := range components {
		if c.Parent == nil {
			roots = append(roots, c)
		} else {
			children[*c.Parent] = append(children[*c.Parent], c)
		}
	}

	var sb strings.Builder
	sb.WriteString("\n# Доступные компоненты:\n")

	var write func(comps []api.Component, indent string)
	write = func(comps []api.Component, indent string) {
		for _, c := range comps {
			sb.WriteString(fmt.Sprintf("#%s- %s\n", indent, c.Name))
			if c.ID != nil {
				if kids, ok := children[*c.ID]; ok {
					write(kids, indent+"  ")
				}
			}
		}
	}
	write(roots, " ")

	return sb.String()
}

func selectComponentsInteractive(components []api.Component) ([]string, error) {
	entries := flattenComponentTree(components)
	selected := make(map[int]string) // entry index -> "impact name"

	impactLabels := make([]string, len(impactChoices))
	for i, ic := range impactChoices {
		impactLabels[i] = ic.label
	}

	for {
		items := make([]string, 0, len(entries)+1)
		for i, e := range entries {
			if impact, ok := selected[i]; ok {
				items = append(items, fmt.Sprintf("[%s] %s", impact, e.display))
			} else {
				items = append(items, fmt.Sprintf("    %s", e.display))
			}
		}
		items = append(items, "Done")

		sel := promptui.Select{
			Label: "Выберите компонент (Enter — добавить/изменить, Done — завершить)",
			Items: items,
			Size:  min(len(items), 12),
		}

		idx, _, err := sel.Run()
		if err != nil {
			return nil, err
		}

		if idx == len(entries) {
			break
		}

		impactItems := append([]string{}, impactLabels...)
		if _, alreadySelected := selected[idx]; alreadySelected {
			impactItems = append(impactItems, "✕  — убрать из списка")
		} else {
			impactItems = append(impactItems, "← Отмена")
		}

		impactSel := promptui.Select{
			Label: fmt.Sprintf("Impact для %q", entries[idx].component.Name),
			Items: impactItems,
		}

		impactIdx, _, err := impactSel.Run()
		if err != nil {
			return nil, err
		}

		if impactIdx == len(impactChoices) {
			delete(selected, idx)
		} else {
			selected[idx] = impactChoices[impactIdx].key + " " + entries[idx].component.Name
		}
	}

	result := make([]string, 0, len(selected))
	for i := range entries {
		if s, ok := selected[i]; ok {
			result = append(result, s)
		}
	}
	return result, nil
}

func validateCreateIncidentPayload(p *api.CreateIncidentPayload) []string {
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
	if p.StartAt.IsZero() {
		errs = append(errs, "start_at: обязательное поле")
	}
	return errs
}
