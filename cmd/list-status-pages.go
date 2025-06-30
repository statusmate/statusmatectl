package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"statusmatectl/api"
	"statusmatectl/printer"
)

var ListStatusPagesCmd = &cobra.Command{
	Use:   "list-status-pages",
	Short: "Ls command",
	RunE:  listStatusPagesCmdF,
}

func init() {
	ListStatusPagesCmd.Flags().String("format", printer.PrintTableFormatTable, "Format output")
	RootCmd.AddCommand(ListStatusPagesCmd)
}

func listStatusPagesCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	payload := api.NewAllPaginatedRequest(nil)
	data, err := client.GetPaginatedStatusPages(payload)
	if err != nil {
		return err
	}

	config := printer.NewPrintTableConfig()
	config.PrintBlockTotal = false

	format, err := command.Flags().GetString("format")
	if err != nil {
		return errors.Wrap(err, "format flag error")
	}

	if err := printer.ValidatePrintTableFormat(format); err != nil {
		return err
	}

	config.Format = format

	err = printer.PrintStatusPages(os.Stdout, data, config)
	if err != nil {
		return err
	}

	return nil
}
