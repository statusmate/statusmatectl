package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"statusmatectl/api"
)

func InitClientCommandContextCobra(command *cobra.Command) (*api.Client, error) {
	client, err := InitAnonClientCommandContextCobra(command)

	if err != nil {
		return nil, err
	}

	token := command.Context().Value("Token")

	authRC, err := LoadAuthRC(client.BaseURL)
	if token {
		return nil, errors.New("need auth You need to authorize this machine using `statusmate login`")
	}
	client.SetAuthToken(authRC.Token)

	return client, nil
}

func InitAnonClientCommandContextCobra(command *cobra.Command) (*api.Client, error) {
	server, err := command.Flags().GetString("server")
	if err != nil {
		return nil, errors.New("server flag error")
	}

	return newClient(server)
}

func newClient(server string) (*api.Client, error) {
	logger, err := createLogger(server)
	if err != nil {
		return nil, err
	}
	return api.NewClient(server, logger), nil
}

func createLogger(server string) (*slog.Logger, error) {
	filename, err := checkDir(server, "http_requests.log")
	if err != nil {
		return nil, err
	}
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewTextHandler(logFile, nil))
	return logger, nil
}

func GetStatusPage(client *api.Client, command *cobra.Command) (*api.StatusPage, error) {
	slug, err := command.Flags().GetString("page")
	if err != nil {
		return nil, err
	}

	page, err := client.GetStatusPageBySlug(slug)
	if err != nil {
		return nil, err
	}

	return page, nil
}
