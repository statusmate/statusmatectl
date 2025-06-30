package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var UpdateCmd = &cobra.Command{
	Use: "update",
}

var UpdateMaintenanceCmd = &cobra.Command{
	Use:     "maintenance <maintenance_id>",
	Aliases: []string{"m"},
	Short:   "Update maintenance",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("Add maintenance")
	},
}

func init() {
	UpdateCmd.AddCommand(
		UpdateMaintenanceCmd,
	)
	RootCmd.AddCommand(UpdateCmd)
}
