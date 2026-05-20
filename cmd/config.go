package cmd

import (
	"encoding/json"
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

var ConfigPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show path to the config file",
	RunE:  configPathCmdF,
}

var ConfigShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current config",
	RunE:  configShowCmdF,
}

func init() {
	ConfigCmd.AddCommand(UseStatusPageCmd)
	ConfigCmd.AddCommand(ConfigPathCmd)
	ConfigCmd.AddCommand(ConfigShowCmd)
	RootCmd.AddCommand(ConfigCmd)
}

func getFirstValue(input string) (string, error) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", errors.New("the string is empty or contains no values")
	}
	return parts[0], nil
}

func configPathCmdF(command *cobra.Command, args []string) error {
	server, err := command.Flags().GetString("server")
	if err != nil {
		return err
	}
	path, err := checkDir(server, "authrc")
	if err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}

func configShowCmdF(command *cobra.Command, args []string) error {
	server, err := command.Flags().GetString("server")
	if err != nil {
		return err
	}
	authRC, err := LoadAuthRC(server)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(authRC, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
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
		items[idx] = fmt.Sprintf("%s %s", statusPage.Slug, statusPage.AbsoluteURL)
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

	slug, err := getFirstValue(result)
	if err != nil {
		return err
	}

	server, err := command.Flags().GetString("server")
	if err != nil {
		return err
	}

	authRC, err := LoadAuthRC(server)
	if err != nil {
		return err
	}

	authRC.DefaultStatusPage = slug

	if err := SaveAuthRC(server, authRC); err != nil {
		return err
	}

	fmt.Printf("Default status page set to %s\n", slug)
	return nil
}
