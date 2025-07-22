package cmd

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:          "st4",
	Short:        "Statusmate cli tool",
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
	RootCmd.PersistentFlags().String("server", "https://statusmate.top", "Server api url")
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
