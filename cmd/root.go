package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

const envServer = "ST4_SERVER"

var RootCmd = &cobra.Command{
	Use:          "st4",
	Short:        "Statusmate cli tool",
	Version:      "dev",
	SilenceUsage: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		checkForRootUser()
	},
}

func SetVersion(v string) {
	RootCmd.Version = v
}

func init() {
	defaultServer := os.Getenv(envServer)
	if defaultServer == "" {
		defaultServer = "statusmate.top"
	}
	RootCmd.PersistentFlags().String("server", defaultServer, "Server api url (env: ST4_SERVER)")
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Show detailed information")
	RootCmd.PersistentFlags().BoolP("pick", "P", false, "Interactively select status page")
}

func checkForRootUser() {
	if os.Geteuid() == 0 {
		slog.Warn("Running Statusmate as root is not recommended. Please use a non-root user.")
	}
}
