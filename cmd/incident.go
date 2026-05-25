package cmd

import (
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/printer"
)

var IncidentCmd = &cobra.Command{
	Use:     "incident",
	Aliases: []string{"inc", "i"},
	Short:   "Manage incidents",
}

// incident ls
var incidentLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List incidents",
	RunE:  listIncidentsCmdF,
}

// incident create
var incidentCreateSubCmd = &cobra.Command{
	Use:   "create",
	Short: "Create incident",
	RunE:  createIncidentCmdF,
}

// incident show [uuid]
var incidentShowSubCmd = &cobra.Command{
	Use:   "show [uuid]",
	Short: "Show incident details",
	Args:  cobra.MaximumNArgs(1),
	RunE:  showIncidentCmdF,
}

// incident close [uuid]
var incidentCloseSubCmd = &cobra.Command{
	Use:   "close [uuid]",
	Short: "Resolve an incident (pick interactively if uuid not given)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  closeIncidentCmdF,
}

// incident touch [file]
var incidentTouchSubCmd = &cobra.Command{
	Use:   "touch [file]",
	Short: "Create an incident template file",
	Args:  cobra.MaximumNArgs(1),
	RunE:  touchIncidentCmdF,
}

// incident open [uuid]
var incidentOpenSubCmd = &cobra.Command{
	Use:   "open [uuid]",
	Short: "Open an incident in the browser",
	Args:  cobra.MaximumNArgs(1),
	RunE:  openIncidentCmdF,
}

func init() {
	// ls flags (same as list-incidents)
	incidentLsCmd.Flags().BoolP("all", "a", false, "List all incidents including resolved")
	incidentLsCmd.Flags().StringP("page", "p", "", "Status page")
	incidentLsCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table, json")

	// create flags (mirrors CreateIncidentCmd)
	incidentCreateSubCmd.Flags().StringP("page", "p", "", "Status page or default")
	incidentCreateSubCmd.Flags().StringP("status", "s", string(api.IncidentStatusInvestigation), "Status")
	incidentCreateSubCmd.Flags().StringP("title", "n", "", "Title of the incident")
	incidentCreateSubCmd.Flags().StringP("desc", "d", "", "Description of the incident")
	incidentCreateSubCmd.Flags().StringArrayP("components", "c", []string{}, "Affected components in 'impact component' format")
	incidentCreateSubCmd.Flags().String("private", "", "Private note")
	incidentCreateSubCmd.Flags().Bool("showOnTop", false, "Show incident on top")
	incidentCreateSubCmd.Flags().Bool("notify", true, "Send notifications")
	incidentCreateSubCmd.Flags().BoolP("interactive", "i", false, "Interactive editing mode")
	incidentCreateSubCmd.Flags().BoolP("pick-components", "C", false, "Interactively pick components")
	incidentCreateSubCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	incidentCreateSubCmd.Flags().Bool("dry", false, "Dry run")
	incidentCreateSubCmd.Flags().StringP("file", "f", "", "Path to incident template file")

	// show flags
	incidentShowSubCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table, json")
	incidentShowSubCmd.Flags().StringP("page", "p", "", "Status page")
	incidentShowSubCmd.Flags().BoolP("all", "a", false, "Include resolved incidents")

	// close flags (mirrors CloseIncidentsCmd)
	incidentCloseSubCmd.Flags().StringP("page", "p", "", "Status page")
	incidentCloseSubCmd.Flags().StringP("message", "m", "", "Resolution message (optional)")
	incidentCloseSubCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	incidentCloseSubCmd.Flags().Bool("dry", false, "Dry run")
	incidentCloseSubCmd.Flags().Bool("notify", true, "Send notifications")

	// touch flags
	incidentTouchSubCmd.Flags().StringP("page", "p", "", "Status page or default")

	// open flags
	incidentOpenSubCmd.Flags().StringP("page", "p", "", "Status page")
	incidentOpenSubCmd.Flags().BoolP("all", "a", false, "Include resolved incidents")

	IncidentCmd.AddCommand(incidentLsCmd)
	IncidentCmd.AddCommand(incidentCreateSubCmd)
	IncidentCmd.AddCommand(incidentShowSubCmd)
	IncidentCmd.AddCommand(incidentCloseSubCmd)
	IncidentCmd.AddCommand(incidentTouchSubCmd)
	IncidentCmd.AddCommand(incidentOpenSubCmd)

	RootCmd.AddCommand(IncidentCmd)
}
