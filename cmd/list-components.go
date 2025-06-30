package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"statusmatectl/api"
	"statusmatectl/printer"
)

var ListComponentsCmd = &cobra.Command{
	Use:   "list-components",
	Short: "Ls command",
	PreRunE:
	RunE:  listComponentsCmdF,
}

func init() {
	ListComponentsCmd.Flags().BoolP("all", "a", false, "List all components")
	ListComponentsCmd.Flags().String("page", "", "Status page")
	ListComponentsCmd.Flags().String("format", printer.PrintTableFormatTable, "Format output")
	ListComponentsCmd.Flags().Bool("total", false, "Total output")

	RootCmd.AddCommand(ListComponentsCmd)
}

func listComponentsCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	statusPage, err := GetStatusPage(client, command)
	if err != nil {
		return err
	}

	showAll, err := command.Flags().GetBool("all")
	if err != nil {
		return errors.Wrap(err, "all flag error")
	}

	total, err := command.Flags().GetBool("total")
	if err != nil {
		return errors.Wrap(err, "total flag error")
	}

	filters := api.PaginatedRequestFilter{
		"status_page": statusPage.ID,
		"enabled":     "true",
	}

	if showAll {
		delete(filters, "enabled")
	}

	payload := api.NewAllPaginatedRequest(filters)
	data, err := client.GetPaginatedComponents(payload)
	if err != nil {
		return err
	}

	config := printer.NewPrintTableConfig()
	config.PrintBlockTotal = total

	format, err := command.Flags().GetString("format")
	if err != nil {
		return errors.Wrap(err, "format flag error")
	}

	if err := printer.ValidatePrintTableFormat(format); err != nil {
		return err
	}

	config.Format = format

	err = printer.PrintComponents(os.Stdout, data, config)
	if err != nil {
		return err
	}

	return nil
}
