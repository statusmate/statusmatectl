package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/printer"
)

// subscriber — корневая команда-неймспейс
var SubscriberCmd = &cobra.Command{
	Use:     "subscriber",
	Aliases: []string{"sub"},
	Short:   "Manage subscribers",
}

var ListSubscribersCmd = &cobra.Command{
	Use:   "list",
	Short: "List subscribers",
	RunE:  listSubscribersCmdF,
}

var ShortListSubscribersCmd = &cobra.Command{
	Use:   "s",
	Short: "List subscribers",
	RunE:  listSubscribersCmdF,
}

var CreateSubscriberCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a subscriber",
	RunE:  createSubscriberCmdF,
}

var DeleteSubscriberCmd = &cobra.Command{
	Use:   "delete <uuid>",
	Short: "Delete a subscriber",
	Args:  cobra.ExactArgs(1),
	RunE:  deleteSubscriberCmdF,
}

var VerifySubscriberCmd = &cobra.Command{
	Use:   "verify <uuid>",
	Short: "Mark a subscriber as verified",
	Args:  cobra.ExactArgs(1),
	RunE:  verifySubscriberCmdF,
}

func init() {
	ListSubscribersCmd.Flags().StringP("page", "p", "", "Status page")
	ListSubscribersCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table|json")
	ListSubscribersCmd.Flags().Bool("email", false, "Filter: email subscribers only")
	ListSubscribersCmd.Flags().Bool("webhook", false, "Filter: webhook subscribers only")

	ShortListSubscribersCmd.Flags().StringP("page", "p", "", "Status page")
	ShortListSubscribersCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table|json")
	ShortListSubscribersCmd.Flags().Bool("email", false, "Filter: email subscribers only")
	ShortListSubscribersCmd.Flags().Bool("webhook", false, "Filter: webhook subscribers only")

	CreateSubscriberCmd.Flags().StringP("page", "p", "", "Status page")
	CreateSubscriberCmd.Flags().StringP("email", "e", "", "Subscriber email (required)")
	CreateSubscriberCmd.Flags().Bool("no-email", false, "Disable email subscription (use with webhook)")
	_ = CreateSubscriberCmd.MarkFlagRequired("email")

	SubscriberCmd.AddCommand(ListSubscribersCmd)
	SubscriberCmd.AddCommand(CreateSubscriberCmd)
	SubscriberCmd.AddCommand(DeleteSubscriberCmd)
	SubscriberCmd.AddCommand(VerifySubscriberCmd)

	RootCmd.AddCommand(SubscriberCmd)
	LsCmd.AddCommand(ShortListSubscribersCmd)
}

func listSubscribersCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	statusPage, err := GetStatusPage(client, command)
	if err != nil {
		return err
	}

	filters := api.PaginatedRequestFilter{
		"status_page": statusPage.ID,
	}

	if onlyEmail, _ := command.Flags().GetBool("email"); onlyEmail {
		filters["subscribe_by_email"] = true
	}
	if onlyWebhook, _ := command.Flags().GetBool("webhook"); onlyWebhook {
		filters["subscribe_by_webhook"] = true
	}

	data, err := client.GetPaginatedSubscribers(api.NewAllPaginatedRequest(filters))
	if err != nil {
		return err
	}

	format, _ := command.Flags().GetString("format")
	if err := printer.ValidatePrintTableFormat(format); err != nil {
		return err
	}

	config := printer.NewPrintTableConfig()
	config.Format = format
	config.PrintBlockTotal = true

	return printer.PrintSubscribers(os.Stdout, data, config)
}

func createSubscriberCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	statusPage, err := GetStatusPage(client, command)
	if err != nil {
		return err
	}

	email, _ := command.Flags().GetString("email")
	noEmail, _ := command.Flags().GetBool("no-email")

	payload := &api.CreateSubscriberPayload{
		Email:            email,
		StatusPage:       statusPage.ID,
		SubscribeByEmail: !noEmail,
	}

	sub, err := client.CreateSubscriber(payload)
	if err != nil {
		return err
	}

	return printer.PrintSummarySubscriber(os.Stdout, sub)
}

func deleteSubscriberCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	uuid := args[0]

	if err := client.DeleteSubscriber(uuid); err != nil {
		return err
	}

	fmt.Printf("Subscriber %s deleted.\n", uuid)
	return nil
}

func verifySubscriberCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	uuid := args[0]

	if err := client.VerifySubscriber(uuid); err != nil {
		return err
	}

	fmt.Printf("Subscriber %s marked as verified.\n", uuid)
	return nil
}
