package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
)

var ConfigCmd = &cobra.Command{
	Use: "config",
}

var UseStatusPageCmd = &cobra.Command{
	Use:   "use-status-page",
	Short: "Use default status page",
	RunE:  useStatusPageCmdF,
}

func init() {
	ConfigCmd.AddCommand(UseStatusPageCmd)
	RootCmd.AddCommand(ConfigCmd)
}

func getFirstValue(input string) (string, error) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", errors.New("the string is empty or contains no values")
	}
	return parts[0], nil
}

func useStatusPageCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	payload := api.NewAllPaginatedRequest(api.PaginatedRequestFilter{})
	statusPages, err := client.GetPaginatedStatusPages(payload)
	if err != nil {
		return err
	}

	items := make([]string, statusPages.Count)

	for idx, statusPage := range statusPages.Results {
		items[idx] = fmt.Sprintf("%s %s", statusPage.UUID, statusPage.AbsoluteURL)
	}

	prompt := promptui.Select{
		Label: "Select Status page",
		Items: items,
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return nil
	}

	UUID, err := getFirstValue(result)

	if err != nil {
		return err
	}

	fmt.Printf("You choose %s\n", UUID)

	// сохранить в authrc
	return nil
}
