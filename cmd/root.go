package cmd

import (
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

var RootCmd = &cobra.Command{
	Use:          "statusmate",
	Short:        "StatusMate CLI tool",
	SilenceUsage: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		checkForRootUser()
	},
}

func init() {
	RootCmd.PersistentFlags().String("server", "https://devstatusmate.ru", "Server system")
}

func checkForRootUser() {
	if os.Geteuid() == 0 {
		slog.Warn("Running Statusmate as root is not recommended. Please use a non-root user.")
	}
}
