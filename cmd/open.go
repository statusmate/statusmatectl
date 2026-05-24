package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
)

var OpenCmd = &cobra.Command{
	Use:   "open",
	Short: "Open a status page in the browser",
	RunE:  openCmdF,
}

func init() {
	OpenCmd.Flags().StringP("page", "p", "", "Status page slug (default: configured default page)")
	OpenCmd.Flags().BoolP("interactive", "i", false, "Select status page interactively")
	RootCmd.AddCommand(OpenCmd)
}

func openCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	interactive, _ := command.Flags().GetBool("interactive")

	var url string
	if interactive {
		url, err = pickStatusPageURL(client)
		if err != nil {
			return err
		}
	} else {
		page, err := GetStatusPage(client, command)
		if err != nil {
			return err
		}
		url = page.AbsoluteURL
	}

	fmt.Println(url)

	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", url)
	} else {
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func pickStatusPageURL(client *api.Client) (string, error) {
	payload := api.NewAllPaginatedRequest(api.PaginatedRequestFilter{})
	statusPages, err := client.GetPaginatedStatusPages(payload)
	if err != nil {
		return "", err
	}

	items := make([]string, len(statusPages.Results))
	for i, sp := range statusPages.Results {
		items[i] = fmt.Sprintf("%s %s", sp.Slug, sp.AbsoluteURL)
	}

	prompt := promptui.Select{
		Label: "Select Status page",
		Items: items,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return statusPages.Results[idx].AbsoluteURL, nil
}
