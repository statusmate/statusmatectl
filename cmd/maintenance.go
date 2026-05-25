package cmd

import (
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/printer"
)

var MaintenanceCmd = &cobra.Command{
	Use:   "maintenance",
	Short: "Manage maintenances",
}

var maintenanceLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List maintenances",
	RunE:  listMaintenancesCmdF,
}

var maintenanceCreateSubCmd = &cobra.Command{
	Use:   "create",
	Short: "Schedule maintenance",
	RunE:  createMaintenanceCmdF,
}

// maintenance show [uuid]
var maintenanceShowSubCmd = &cobra.Command{
	Use:   "show [uuid]",
	Short: "Show maintenance details",
	Args:  cobra.MaximumNArgs(1),
	RunE:  showMaintenanceCmdF,
}

// maintenance close [uuid]
var maintenanceCloseSubCmd = &cobra.Command{
	Use:   "close [uuid]",
	Short: "Complete a maintenance (pick interactively if uuid not given)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  closeMaintenanceCmdF,
}

// maintenance touch [file]
var maintenanceTouchSubCmd = &cobra.Command{
	Use:   "touch [file]",
	Short: "Create a maintenance template file",
	Args:  cobra.MaximumNArgs(1),
	RunE:  touchMaintenanceCmdF,
}

// maintenance open [uuid]
var maintenanceOpenSubCmd = &cobra.Command{
	Use:   "open [uuid]",
	Short: "Open a maintenance in the browser",
	Args:  cobra.MaximumNArgs(1),
	RunE:  openMaintenanceCmdF,
}

func init() {
	maintenanceLsCmd.Flags().BoolP("all", "a", false, "List all maintenances including completed")
	maintenanceLsCmd.Flags().StringP("page", "p", "", "Status page")
	maintenanceLsCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table, json")

	maintenanceCreateSubCmd.Flags().StringP("page", "p", "", "Status page or default")
	maintenanceCreateSubCmd.Flags().StringP("title", "n", "", "Title of the maintenance")
	maintenanceCreateSubCmd.Flags().StringP("desc", "d", "", "Description")
	maintenanceCreateSubCmd.Flags().String("start-at", "", "Start time (RFC3339)")
	maintenanceCreateSubCmd.Flags().String("end-at", "", "End time (RFC3339, optional)")
	maintenanceCreateSubCmd.Flags().StringArrayP("components", "c", []string{}, "Affected components in 'impact component' format")
	maintenanceCreateSubCmd.Flags().Bool("notify", true, "Send notifications to subscribers")
	maintenanceCreateSubCmd.Flags().Bool("auto-start", false, "Auto-start at start_at time")
	maintenanceCreateSubCmd.Flags().Bool("auto-end", false, "Auto-end at end_at time")
	maintenanceCreateSubCmd.Flags().Bool("affect-uptime", true, "Count against uptime")
	maintenanceCreateSubCmd.Flags().BoolP("interactive", "i", false, "Interactive editing mode")
	maintenanceCreateSubCmd.Flags().BoolP("pick-components", "C", false, "Interactively pick components")
	maintenanceCreateSubCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	maintenanceCreateSubCmd.Flags().Bool("dry", false, "Dry run")
	maintenanceCreateSubCmd.Flags().StringP("file", "f", "", "Path to maintenance template file (created with touch)")

	maintenanceShowSubCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table, json")
	maintenanceShowSubCmd.Flags().StringP("page", "p", "", "Status page")
	maintenanceShowSubCmd.Flags().BoolP("all", "a", false, "Include completed maintenances")

	maintenanceCloseSubCmd.Flags().StringP("page", "p", "", "Status page")
	maintenanceCloseSubCmd.Flags().StringP("message", "m", "", "Completion message (optional)")
	maintenanceCloseSubCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	maintenanceCloseSubCmd.Flags().Bool("dry", false, "Dry run")
	maintenanceCloseSubCmd.Flags().Bool("notify", true, "Send notifications")

	maintenanceTouchSubCmd.Flags().StringP("page", "p", "", "Status page or default")

	maintenanceOpenSubCmd.Flags().StringP("page", "p", "", "Status page")
	maintenanceOpenSubCmd.Flags().BoolP("all", "a", false, "Include completed maintenances")

	MaintenanceCmd.AddCommand(maintenanceLsCmd)
	MaintenanceCmd.AddCommand(maintenanceCreateSubCmd)
	MaintenanceCmd.AddCommand(maintenanceShowSubCmd)
	MaintenanceCmd.AddCommand(maintenanceCloseSubCmd)
	MaintenanceCmd.AddCommand(maintenanceTouchSubCmd)
	MaintenanceCmd.AddCommand(maintenanceOpenSubCmd)

	RootCmd.AddCommand(MaintenanceCmd)
}

