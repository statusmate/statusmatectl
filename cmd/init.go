package cmd

import (
	"errors"
	"io"
	"log/slog"
	"os"

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
	client.SetAuthToken(authRC.Token)
	client.SetStatusPage(authRC.DefaultStatusPage)
	client.SetReleasePage(authRC.DefaultReleasePage)

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
	slug, err := command.Flags().GetString("page")
	if err != nil {
		return nil, err
	}

	if slug == "" {
		slug = cl.StatusPage
	}

	if slug == "" {
		return nil, errors.New("no status page specified: use --page or set default with `st4 config use-status-page`")
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
