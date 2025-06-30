package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log/slog"
	"net/url"
	"os"
)

var RootCmd = &cobra.Command{
	Use:          "statusmate",
	Short:        "StatusMate CLI tool",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		server, err := cmd.Flags().GetString("server")
		if err != nil {
			return err
		}

		return validateServer(server)
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		checkForRootUser()
	},
}

func init() {
	RootCmd.PersistentFlags().String("server", "https://devstatusmate.ru", "Server system")
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Show detailed information")
}

func validateServer(u string) error {
	if !(u[:7] == "http://" || u[:8] == "https://") {
		return fmt.Errorf("invalid config: must be full URL with \"http://\" or \"https://\"")
	}

	_, err := url.ParseRequestURI(u)
	if err != nil {
		return fmt.Errorf("invalid URL: %s", u)
	}
	return nil
}

func checkForRootUser() {
	if os.Geteuid() == 0 {
		slog.Warn("Running Statusmate as root is not recommended. Please use a non-root user.")
	}
}
