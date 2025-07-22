package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/editor"
)

var CreateIncidentCmd = &cobra.Command{
	Use:   "create-incident",
	Short: "Create incident or maintenance",
	RunE:  createIncidentCmdF,
}

/*
*
Output:

Incident "Database Outage" created successfully!

Summary:
uuid=12345
name=Database Outage
description=All DB servers are down
components=DB, API
status=Investigating
created_at=2024-10-22 12:34

Такой ответ будет удобно читать через awk
| awk -F= '/uuid/ {print $2}'
*/
func init() {
	//statusmate create-incident --page <page_id>
	CreateIncidentCmd.Flags().StringP("page", "p", "", "Status page or default")
	//statusmate create-incident --status investigation
	CreateIncidentCmd.Flags().StringP("status", "s", string(api.IncidentStatusInvestigation), "Status")
	//statusmate create-incident --name "Проблема"
	CreateIncidentCmd.Flags().StringP("name", "n", "", "Name of the incident or maintenance")
	//statusmate create-incident --desc "Проблема"
	CreateIncidentCmd.Flags().StringP("desc", "d", "", "Description of the incident or maintenance")
	//statusmate create-incident  --components="o cdn" --components="p web"
	CreateIncidentCmd.Flags().StringArrayP("components", "c", []string{}, "Specify components with impact in 'impact component' format")
	//statusmate create-incident  --private="Private message"
	CreateIncidentCmd.Flags().String("private", "", "Affected components, e.g. op cloud/lkk")
	//statusmate create-incident --endAt="2024-10-22T14:30:00"
	CreateIncidentCmd.Flags().String("endAt", "", "2024-10-22T14:30:00")
	//statusmate create-incident  --showOnTop
	CreateIncidentCmd.Flags().Bool("showOnTop", false, "Affected components, e.g. op cloud/lkk")
	//statusmate create-incident  --notify
	CreateIncidentCmd.Flags().Bool("notify", true, "Send notify")
	//statusmate create-incident  -y
	CreateIncidentCmd.Flags().BoolP("yes", "y", false, "yes for prompt")
	//statusmate create-incident  --dry
	CreateIncidentCmd.Flags().Bool("dry", false, "Dry run, check data and open editor from $EDITOR")

	RootCmd.AddCommand(CreateIncidentCmd)
}

func createIncidentCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	name, err := command.Flags().GetString("name")
	if err != nil {
		return errors.New("import-path flag error")
	}

	componentImpacts, err := command.Flags().GetStringArray("components")
	if err != nil {
		return errors.New("components flag error")
	}

	var parsedComponentImpacts []api.ComponentImpact

	for _, ci := range componentImpacts {
		parsedCI, err := api.ParseComponentImpact(ci)
		if err != nil {
			return err
		}
		parsedComponentImpacts = append(parsedComponentImpacts, parsedCI)
	}

	for _, ci := range parsedComponentImpacts {
		fmt.Printf("Component: %s, Impact: %s\n", ci.Component, ci.Impact)
	}

	fmt.Printf("Name: %s \n", name)

	statusPage, err := GetStatusPage(client, command)
	if err != nil {
		return err
	}

	payload := api.NewAllPaginatedRequest(api.PaginatedRequestFilter{
		"status_page": statusPage.ID,
	})

	components, err := client.GetPaginatedComponents(payload)
	if err != nil {
		return err
	}

	fmt.Printf("%v", components)

	var inputReader io.Reader = command.InOrStdin()

	if len(args) > 0 && args[0] != "-" {
		file, err := os.Open(args[0])

		if err != nil {
			return fmt.Errorf("failed open file: %v", err)
		}

		inputReader = file
	}

	data, err := io.ReadAll(inputReader)
	if err != nil {
		return err
	}

	output, err := editor.CaptureInputFromEditor(data, editor.GetPreferredEditorFromEnvironment)
	if err != nil {
		return err
	}

	fmt.Printf("%s \n", string(output))

	return nil
}
