package cmd

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
)

func InitClientCommandContextCobra(command *cobra.Command) (*api.Client, error) {
	server, err := command.Flags().GetString("server")
	if err != nil {
		return nil, errors.New("server flag error")
	}
	verbose, _ := command.Flags().GetBool("verbose")

	client, err := NewClient(server, verbose)
	if err != nil {
		return nil, err
	}

	authRC, err := LoadAuthRC(client.BaseURL)
	if err != nil {
		return nil, errors.New("need auth You need to authorize this machine using `st4 login`")
	}
	client.AuthRC = authRC

	return client, nil
}

func InitAnonClientCommandContextCobra(command *cobra.Command) (*api.Client, error) {
	server, err := command.Flags().GetString("server")
	if err != nil {
		return nil, errors.New("server flag error")
	}
	verbose, _ := command.Flags().GetBool("verbose")

	return NewClient(server, verbose)
}

func NewClient(server string, verbose bool) (*api.Client, error) {
	logger, err := createLogger(server, verbose)
	if err != nil {
		return nil, err
	}
	return api.NewClient(server, logger), nil
}

func createLogger(server string, verbose bool) (*slog.Logger, error) {
	filename, err := checkDir(server, "http_requests.log")
	if err != nil {
		return nil, err
	}
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	var w io.Writer = logFile
	if verbose {
		w = io.MultiWriter(logFile, os.Stderr)
	}

	logger := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return logger, nil
}

func GetStatusPage(cl *api.Client, command *cobra.Command) (*api.StatusPage, error) {
	pick, _ := command.Root().PersistentFlags().GetBool("pick")
	if pick {
		return pickStatusPage(cl)
	}

	slug, err := command.Flags().GetString("page")
	if err != nil {
		return nil, err
	}

	if slug == "" && cl.AuthRC != nil {
		slug = cl.AuthRC.DefaultStatusPage
	}

	if slug == "" {
		return pickStatusPage(cl)
	}

	if IdentifyType(slug) == TypeUUID {
		return getStatusPageByUUID(cl, slug)
	}

	return cl.GetStatusPageBySlug(slug)
}

func getStatusPageByUUID(cl *api.Client, uuid string) (*api.StatusPage, error) {
	pages, err := cl.GetPaginatedStatusPages(api.NewAllPaginatedRequest(api.PaginatedRequestFilter{}))
	if err != nil {
		return nil, err
	}
	for _, p := range pages.Results {
		if p.UUID == uuid {
			return &p, nil
		}
	}
	return nil, errors.New("status page not found: " + uuid)
}

func pickStatusPage(cl *api.Client) (*api.StatusPage, error) {
	pages, err := cl.GetPaginatedStatusPages(api.NewAllPaginatedRequest(api.PaginatedRequestFilter{}))
	if err != nil {
		return nil, err
	}
	if len(pages.Results) == 0 {
		return nil, errors.New("no status pages found")
	}

	items := make([]string, len(pages.Results))
	for i, p := range pages.Results {
		items[i] = fmt.Sprintf("%s  %s", p.Slug, p.AbsoluteURL)
	}

	prompt := promptui.Select{
		Label: "Select status page",
		Items: items,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	selected := &pages.Results[idx]

	if cl.AuthRC != nil {
		cl.AuthRC.DefaultStatusPage = selected.Slug
		_ = SaveAuthRC(cl.BaseURL, cl.AuthRC)
	}

	return selected, nil
}
