package cmd

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/printer"
)

var TeamCmd = &cobra.Command{
	Use:     "team",
	Aliases: []string{"teams"},
	Short:   "Manage teams",
}

var ListTeamsCmd = &cobra.Command{
	Use:   "list",
	Short: "List teams",
	RunE:  listTeamsCmdF,
}

var ShortListTeamsCmd = &cobra.Command{
	Use:   "t",
	Short: "List teams",
	RunE:  listTeamsCmdF,
}

var InviteUserCmd = &cobra.Command{
	Use:   "invite",
	Short: "Invite a user to a team",
	RunE:  inviteUserCmdF,
}

var ListTeamInvitesCmd = &cobra.Command{
	Use:   "list-invites",
	Short: "List pending team invites",
	RunE:  listTeamInvitesCmdF,
}

var ListTeamMembersCmd = &cobra.Command{
	Use:   "members",
	Short: "List team members",
	RunE:  listTeamMembersCmdF,
}

var ShortListTeamMembersCmd = &cobra.Command{
	Use:   "u",
	Short: "List team members",
	RunE:  listTeamMembersCmdF,
}

var DeleteTeamInviteCmd = &cobra.Command{
	Use:   "revoke <code>",
	Short: "Revoke a team invite by code",
	Args:  cobra.ExactArgs(1),
	RunE:  deleteTeamInviteCmdF,
}

func init() {
	ListTeamsCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table|json")

	ShortListTeamsCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table|json")

	ListTeamMembersCmd.Flags().Int("team", 0, "Team ID (auto-detected if only one team)")
	ListTeamMembersCmd.Flags().StringP("role", "r", "", "Filter by role: owner|manager")
	ListTeamMembersCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table|json")

	ShortListTeamMembersCmd.Flags().Int("team", 0, "Team ID (auto-detected if only one team)")
	ShortListTeamMembersCmd.Flags().StringP("role", "r", "", "Filter by role: owner|manager")
	ShortListTeamMembersCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table|json")

	InviteUserCmd.Flags().StringP("email", "e", "", "Email of user to invite (required)")
	InviteUserCmd.Flags().StringP("role", "r", "manager", "Role: owner|manager")
	InviteUserCmd.Flags().Int("team", 0, "Team ID (auto-detected if only one team)")
	_ = InviteUserCmd.MarkFlagRequired("email")

	ListTeamInvitesCmd.Flags().Int("team", 0, "Team ID (auto-detected if only one team)")
	ListTeamInvitesCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table|json")

	DeleteTeamInviteCmd.Flags().Int("team", 0, "Team ID (unused, for consistency)")

	TeamCmd.AddCommand(ListTeamsCmd)
	TeamCmd.AddCommand(ListTeamMembersCmd)
	TeamCmd.AddCommand(InviteUserCmd)
	TeamCmd.AddCommand(ListTeamInvitesCmd)
	TeamCmd.AddCommand(DeleteTeamInviteCmd)

	RootCmd.AddCommand(TeamCmd)
	LsCmd.AddCommand(ShortListTeamsCmd)
	LsCmd.AddCommand(ShortListTeamMembersCmd)
}

// getTeamID resolves team ID: --team flag → AuthRC default → single team → promptui picker.
func getTeamID(client *api.Client, command *cobra.Command) (int, error) {
	if command.Flags().Changed("team") {
		id, _ := command.Flags().GetInt("team")
		return id, nil
	}

	if client.AuthRC != nil && client.AuthRC.DefaultTeam != 0 {
		return client.AuthRC.DefaultTeam, nil
	}

	return pickTeam(client)
}

func pickTeam(client *api.Client) (int, error) {
	teams, err := client.GetPaginatedTeams(api.NewAllPaginatedRequest(nil))
	if err != nil {
		return 0, err
	}

	if len(teams.Results) == 0 {
		return 0, fmt.Errorf("no teams found")
	}

	if len(teams.Results) == 1 {
		return teams.Results[0].ID, nil
	}

	items := make([]string, len(teams.Results))
	for i, t := range teams.Results {
		items[i] = fmt.Sprintf("%d  %s", t.ID, t.Name)
	}

	prompt := promptui.Select{
		Label: "Select team",
		Items: items,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return 0, err
	}

	selected := teams.Results[idx]

	if client.AuthRC != nil {
		client.AuthRC.DefaultTeam = selected.ID
		_ = SaveAuthRC(client.BaseURL, client.AuthRC)
	}

	return selected.ID, nil
}

func listTeamsCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	data, err := client.GetPaginatedTeams(api.NewAllPaginatedRequest(nil))
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

	return printer.PrintTeams(os.Stdout, data, config)
}

func inviteUserCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	teamID, err := getTeamID(client, command)
	if err != nil {
		return err
	}

	email, _ := command.Flags().GetString("email")
	role, _ := command.Flags().GetString("role")

	payload := &api.CreateTeamInvitePayload{
		Email: email,
		Role:  role,
		Team:  teamID,
	}

	invite, err := client.CreateTeamInvite(payload)
	if err != nil {
		return err
	}

	return printer.PrintSummaryTeamInvite(os.Stdout, invite)
}

func listTeamInvitesCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	teamID, err := getTeamID(client, command)
	if err != nil {
		return err
	}

	filters := api.PaginatedRequestFilter{
		"team": teamID,
	}

	data, err := client.GetPaginatedTeamInvites(api.NewAllPaginatedRequest(filters))
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

	return printer.PrintTeamInvites(os.Stdout, data, config)
}

func listTeamMembersCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	teamID, err := getTeamID(client, command)
	if err != nil {
		return err
	}

	filters := api.PaginatedRequestFilter{
		"team": teamID,
	}

	if role, _ := command.Flags().GetString("role"); role != "" {
		filters["role"] = role
	}

	data, err := client.GetPaginatedTeamUsersExpanded(api.NewAllPaginatedRequest(filters))
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

	return printer.PrintTeamUsers(os.Stdout, data, config)
}

func deleteTeamInviteCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	code := args[0]

	if err := client.DeleteTeamInvite(code); err != nil {
		return err
	}

	fmt.Printf("Invite %s revoked.\n", code)
	return nil
}
