package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"statusmatectl/editor"
)

// statusmate add i
var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add incident or maintenance",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Run add")
	},
}

var AddIncidentCmd = &cobra.Command{
	Use:     "incident [file]",
	Aliases: []string{"i"},
	Short:   "Add incident",
	RunE:    addIncidentCmdF,
}

var AddMaintenanceCmd = &cobra.Command{
	Use:     "maintenance [file]",
	Aliases: []string{"m"},
	Short:   "Add smaintenance",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("Add maintenance")
	},
}

func init() {
	AddIncidentCmd.Flags().StringP("status", "s", "", "Status")
	AddIncidentCmd.Flags().StringP("page", "p", "", "Status page or default")
	AddIncidentCmd.Flags().StringP("name", "n", "", "Name of the incident or maintenance")
	AddIncidentCmd.Flags().StringP("desc", "d", "", "Description of the incident or maintenance")
	AddIncidentCmd.Flags().StringP("components", "c", "", "Affected components, e.g. op cloud/lkk")
	AddIncidentCmd.Flags().Bool("notify", true, "Send notify")
	AddIncidentCmd.Flags().BoolP("yes", "y", false, "yes for prompt")
	AddIncidentCmd.Flags().Bool("dry", false, "Dry run, check data and open editor from $EDITOR")

	AddCmd.AddCommand(
		AddIncidentCmd,
		AddMaintenanceCmd,
	)

	RootCmd.AddCommand(AddCmd)
}

func addIncidentCmdF(command *cobra.Command, args []string) error {
	name, err := command.Flags().GetString("name")
	if err != nil {
		return errors.New("import-path flag error")
	}

	fmt.Printf("use name %s \n", name)

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
