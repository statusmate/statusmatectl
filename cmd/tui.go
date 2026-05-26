package cmd

import (
	"errors"
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI dashboard",
	RunE:  runTUI,
}

func init() {
	tuiCmd.Flags().StringP("page", "p", "", "Status page slug")
	RootCmd.AddCommand(tuiCmd)
	// Launch TUI when st4 is invoked with no subcommand
	RootCmd.RunE = rootRunTUI
}

func rootRunTUI(cmd *cobra.Command, _ []string) error {
	client, err := InitClientCommandContextCobra(cmd)
	if err != nil {
		return err
	}
	sp, err := loadStatusPageForTUI(client)
	if err != nil {
		return err
	}
	return tui.NewApp(client, sp).Run()
}

func runTUI(cmd *cobra.Command, _ []string) error {
	client, err := InitClientCommandContextCobra(cmd)
	if err != nil {
		return err
	}
	sp, err := GetStatusPage(client, cmd)
	if err != nil {
		return err
	}
	return tui.NewApp(client, sp).Run()
}

// loadStatusPageForTUI resolves the status page for the TUI without requiring
// a --page cobra flag. Uses the saved default, picks automatically if there's
// only one, or prompts interactively.
func loadStatusPageForTUI(client *api.Client) (*api.StatusPage, error) {
	if client.AuthRC != nil && client.AuthRC.DefaultStatusPage != "" {
		return client.GetStatusPageBySlug(client.AuthRC.DefaultStatusPage)
	}

	pages, err := client.GetPaginatedStatusPages(
		api.NewAllPaginatedRequest(api.PaginatedRequestFilter{}),
	)
	if err != nil {
		return nil, err
	}
	if len(pages.Results) == 0 {
		return nil, errors.New("no status pages found")
	}
	if len(pages.Results) == 1 {
		return &pages.Results[0], nil
	}

	items := make([]string, len(pages.Results))
	for i, p := range pages.Results {
		items[i] = fmt.Sprintf("%s  %s", p.Slug, p.AbsoluteURL)
	}
	prompt := promptui.Select{Label: "Select status page", Items: items}
	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	selected := &pages.Results[idx]
	if client.AuthRC != nil {
		client.AuthRC.DefaultStatusPage = selected.Slug
		_ = SaveAuthRC(client.BaseURL, client.AuthRC)
	}
	return selected, nil
}
