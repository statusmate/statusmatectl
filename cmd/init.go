package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"statusmatectl/api"
)

func InitClientCommandContextCobra(command *cobra.Command) (*api.Client, error) {
	server, err := command.Flags().GetString("server")
	if err != nil {
		return nil, errors.New("server flag error")
	}

	if command.Use == "login" {
		return api.NewClient(server), nil
	}
	return api.NewClientWithToken(server)
}

func InitAnonClientCommandContextCobra(command *cobra.Command) (*api.Client, error) {
	server, err := command.Flags().GetString("server")
	if err != nil {
		return nil, errors.New("server flag error")
	}

	return api.NewClient(server), nil
}
