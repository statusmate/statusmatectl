package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/printer"
)

var ComponentCmd = &cobra.Command{
	Use:   "component",
	Short: "Manage components",
}

var componentLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List components",
	RunE:  listComponentsCmdF,
}

var componentShowCmd = &cobra.Command{
	Use:   "show [name|uuid]",
	Short: "Show component details",
	Args:  cobra.MaximumNArgs(1),
	RunE:  componentShowCmdF,
}

var componentCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a component",
	RunE:  componentCreateCmdF,
}

var componentUpdateCmd = &cobra.Command{
	Use:   "update [name|uuid]",
	Short: "Update a component",
	Args:  cobra.MaximumNArgs(1),
	RunE:  componentUpdateCmdF,
}

var componentDeleteCmd = &cobra.Command{
	Use:   "delete [name|uuid]",
	Short: "Delete a component",
	Args:  cobra.MaximumNArgs(1),
	RunE:  componentDeleteCmdF,
}

// Quick status commands — no flags required, just a name or interactive picker.

var componentUpCmd = &cobra.Command{
	Use:   "up [name|uuid]",
	Short: "Set component status to operational",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return componentQuickStatus(cmd, args, api.ImpactTypeOperational)
	},
}

var componentDownCmd = &cobra.Command{
	Use:   "down [name|uuid]",
	Short: "Set component status to major_outage",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return componentQuickStatus(cmd, args, api.ImpactTypeMajorOutage)
	},
}

var componentWarnCmd = &cobra.Command{
	Use:   "warn [name|uuid]",
	Short: "Set component status to degraded_performance",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return componentQuickStatus(cmd, args, api.ImpactTypeDegradedPerformance)
	},
}

var componentPartialCmd = &cobra.Command{
	Use:   "partial [name|uuid]",
	Short: "Set component status to partial_outage",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return componentQuickStatus(cmd, args, api.ImpactTypePartialOutage)
	},
}

var componentEnableCmd = &cobra.Command{
	Use:   "enable [name|uuid]",
	Short: "Enable a component",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return componentSetEnabled(cmd, args, true)
	},
}

var componentDisableCmd = &cobra.Command{
	Use:   "disable [name|uuid]",
	Short: "Disable (hide) a component",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return componentSetEnabled(cmd, args, false)
	},
}

func init() {
	// ls
	componentLsCmd.Flags().BoolP("all", "a", false, "Include disabled components")
	componentLsCmd.Flags().String("page", "", "Status page")
	componentLsCmd.Flags().Bool("total", false, "Print total count")
	componentLsCmd.Flags().String("format", printer.PrintTableFormatTable, "Format: table|list|json")

	// show
	componentShowCmd.Flags().StringP("page", "p", "", "Status page")
	componentShowCmd.Flags().BoolP("all", "a", false, "Include disabled")
	componentShowCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table|json")

	// create
	componentCreateCmd.Flags().StringP("name", "n", "", "Component name (required)")
	componentCreateCmd.Flags().StringP("status", "s", string(api.ImpactTypeOperational), "Initial status (o/d/p/m/u or full name)")
	componentCreateCmd.Flags().StringP("desc", "d", "", "Description")
	componentCreateCmd.Flags().StringP("page", "p", "", "Status page")
	componentCreateCmd.Flags().Bool("histogram", false, "Show uptime histogram")
	componentCreateCmd.Flags().Bool("private", false, "Private component")
	_ = componentCreateCmd.MarkFlagRequired("name")

	// update
	componentUpdateCmd.Flags().StringP("status", "s", "", "Set status (o/d/p/m/u or full name)")
	componentUpdateCmd.Flags().StringP("name", "n", "", "Rename component")
	componentUpdateCmd.Flags().StringP("desc", "d", "", "Update description")
	componentUpdateCmd.Flags().StringP("page", "p", "", "Status page")
	componentUpdateCmd.Flags().BoolP("all", "a", false, "Include disabled in picker")
	componentUpdateCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	// delete
	componentDeleteCmd.Flags().StringP("page", "p", "", "Status page")
	componentDeleteCmd.Flags().BoolP("all", "a", false, "Include disabled in picker")
	componentDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	// quick commands
	for _, cmd := range []*cobra.Command{
		componentUpCmd, componentDownCmd, componentWarnCmd, componentPartialCmd,
		componentEnableCmd, componentDisableCmd,
	} {
		cmd.Flags().StringP("page", "p", "", "Status page")
		cmd.Flags().BoolP("all", "a", false, "Include disabled in picker")
	}

	ComponentCmd.AddCommand(componentLsCmd)
	ComponentCmd.AddCommand(componentShowCmd)
	ComponentCmd.AddCommand(componentCreateCmd)
	ComponentCmd.AddCommand(componentUpdateCmd)
	ComponentCmd.AddCommand(componentDeleteCmd)
	ComponentCmd.AddCommand(componentUpCmd)
	ComponentCmd.AddCommand(componentDownCmd)
	ComponentCmd.AddCommand(componentWarnCmd)
	ComponentCmd.AddCommand(componentPartialCmd)
	ComponentCmd.AddCommand(componentEnableCmd)
	ComponentCmd.AddCommand(componentDisableCmd)

	RootCmd.AddCommand(ComponentCmd)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// resolveComponent finds a component by UUID (exact) or by name (case-insensitive,
// partial match triggers interactive picker when multiple results).
func resolveComponent(client *api.Client, command *cobra.Command, nameOrUUID string) (*api.Component, error) {
	if isUUID(nameOrUUID) {
		return client.GetComponentByUUID(nameOrUUID)
	}

	statusPage, err := GetStatusPage(client, command)
	if err != nil {
		return nil, err
	}

	showAll, _ := command.Flags().GetBool("all")
	filters := api.PaginatedRequestFilter{"status_page": statusPage.ID}
	if !showAll {
		filters["enabled"] = "true"
	}

	comps, err := client.GetPaginatedComponents(api.NewAllPaginatedRequest(filters))
	if err != nil {
		return nil, err
	}

	// exact case-insensitive match first
	for i := range comps.Results {
		if strings.EqualFold(comps.Results[i].Name, nameOrUUID) {
			return &comps.Results[i], nil
		}
	}

	// partial match
	lower := strings.ToLower(nameOrUUID)
	var matches []api.Component
	for _, c := range comps.Results {
		if strings.Contains(strings.ToLower(c.Name), lower) {
			matches = append(matches, c)
		}
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("component not found: %q", nameOrUUID)
	case 1:
		return &matches[0], nil
	default:
		return pickComponent(matches)
	}
}

// resolveComponentOrPick resolves from args or shows an interactive picker.
func resolveComponentOrPick(client *api.Client, command *cobra.Command, args []string) (*api.Component, error) {
	if len(args) == 1 {
		return resolveComponent(client, command, args[0])
	}

	statusPage, err := GetStatusPage(client, command)
	if err != nil {
		return nil, err
	}

	showAll, _ := command.Flags().GetBool("all")
	filters := api.PaginatedRequestFilter{"status_page": statusPage.ID}
	if !showAll {
		filters["enabled"] = "true"
	}

	comps, err := client.GetPaginatedComponents(api.NewAllPaginatedRequest(filters))
	if err != nil {
		return nil, err
	}
	if comps.Count == 0 {
		return nil, fmt.Errorf("no components found")
	}

	return pickComponent(comps.Results)
}

func pickComponent(components []api.Component) (*api.Component, error) {
	items := make([]string, len(components))
	for i, c := range components {
		uuid := ""
		if c.UUID != nil {
			uuid = *c.UUID
		}
		items[i] = fmt.Sprintf("[%s] %s (%s)", c.Impact, c.Name, uuid)
	}

	prompt := promptui.Select{
		Label: "Select component",
		Items: items,
		Size:  min(len(items), 12),
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &components[idx], nil
}

// ── subcommand implementations ────────────────────────────────────────────────

func componentShowCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	format, _ := command.Flags().GetString("format")
	if err := printer.ValidatePrintTableFormat(format); err != nil {
		return err
	}

	comp, err := resolveComponentOrPick(client, command, args)
	if err != nil {
		return err
	}

	return printer.PrintDetailComponent(os.Stdout, comp, format)
}

func componentCreateCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	statusPage, err := GetStatusPage(client, command)
	if err != nil {
		return err
	}

	name, _ := command.Flags().GetString("name")
	statusStr, _ := command.Flags().GetString("status")
	desc, _ := command.Flags().GetString("desc")
	histogram, _ := command.Flags().GetBool("histogram")
	private, _ := command.Flags().GetBool("private")

	impact, err := api.ParseImpact(statusStr)
	if err != nil {
		return err
	}

	comp := &api.Component{
		Name:        name,
		Impact:      impact,
		Description: desc,
		StatusPage:  statusPage.ID,
		Enabled:     true,
		Histogram:   histogram,
		Private:     private,
	}

	created, err := client.CreateComponent(comp)
	if err != nil {
		return err
	}

	printer.PrintSummaryComponent(os.Stdout, created)
	return nil
}

func componentUpdateCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	comp, err := resolveComponentOrPick(client, command, args)
	if err != nil {
		return err
	}

	changed := false

	if command.Flags().Changed("status") {
		statusStr, _ := command.Flags().GetString("status")
		impact, err := api.ParseImpact(statusStr)
		if err != nil {
			return err
		}
		comp.Impact = impact
		changed = true
	}

	if command.Flags().Changed("name") {
		comp.Name, _ = command.Flags().GetString("name")
		changed = true
	}

	if command.Flags().Changed("desc") {
		comp.Description, _ = command.Flags().GetString("desc")
		changed = true
	}

	if !changed {
		return fmt.Errorf("no changes specified; use --status, --name, or --desc")
	}

	yes, _ := command.Flags().GetBool("yes")
	if !yes {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Update component %q?", comp.Name),
			IsConfirm: true,
		}
		if _, err := prompt.Run(); err != nil {
			fmt.Println("Canceled.")
			return nil
		}
	}

	uuid := ""
	if comp.UUID != nil {
		uuid = *comp.UUID
	}

	updated, err := client.UpdateComponent(uuid, comp)
	if err != nil {
		return err
	}

	printer.PrintSummaryComponent(os.Stdout, updated)
	return nil
}

func componentDeleteCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	comp, err := resolveComponentOrPick(client, command, args)
	if err != nil {
		return err
	}

	yes, _ := command.Flags().GetBool("yes")
	if !yes {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Delete component %q? This cannot be undone", comp.Name),
			IsConfirm: true,
		}
		if _, err := prompt.Run(); err != nil {
			fmt.Println("Canceled.")
			return nil
		}
	}

	uuid := ""
	if comp.UUID != nil {
		uuid = *comp.UUID
	}

	if err := client.DeleteComponent(uuid); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "deleted=%s\nuuid=%s\n", comp.Name, uuid)
	return nil
}

func componentSetEnabled(command *cobra.Command, args []string, enabled bool) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	comp, err := resolveComponentOrPick(client, command, args)
	if err != nil {
		return err
	}

	comp.Enabled = enabled

	uuid := ""
	if comp.UUID != nil {
		uuid = *comp.UUID
	}

	updated, err := client.UpdateComponent(uuid, comp)
	if err != nil {
		return err
	}

	printer.PrintSummaryComponent(os.Stdout, updated)
	return nil
}

func componentQuickStatus(command *cobra.Command, args []string, impact api.ImpactType) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	comp, err := resolveComponentOrPick(client, command, args)
	if err != nil {
		return err
	}

	comp.Impact = impact

	uuid := ""
	if comp.UUID != nil {
		uuid = *comp.UUID
	}

	updated, err := client.UpdateComponent(uuid, comp)
	if err != nil {
		return err
	}

	printer.PrintSummaryComponent(os.Stdout, updated)
	return nil
}
