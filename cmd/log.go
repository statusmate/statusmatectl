package cmd

import (
	"fmt"
	"os"
	"time"

	naturaldate "github.com/tj/go-naturaldate"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/printer"
)

var LogCmd = &cobra.Command{
	Use:   "log [COMPONENT_UUID]",
	Short: "Timeline of incidents and maintenances for a component",
	Args:  cobra.MaximumNArgs(1),
	RunE:  logCmdF,
}

func init() {
	LogCmd.Flags().StringP("page", "p", "", "Status page slug or UUID")
	LogCmd.Flags().String("type", "", "Filter by type: incident, maintenance")
	LogCmd.Flags().Int("limit", 0, "Maximum entries to show (0 = all)")
	LogCmd.Flags().String("since", "", "Show events from this time (e.g. yesterday, last week, 2 days ago, 2026-05-20)")
	LogCmd.Flags().String("until", "", "Show events until this time (same formats as --since)")
	LogCmd.Flags().String("format", printer.PrintTableFormatTable, "Output format: table, json, timeline")
	RootCmd.AddCommand(LogCmd)
}

func logCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	statusPage, err := GetStatusPage(client, command)
	if err != nil {
		return errors.Wrap(err, "page flag error")
	}

	format, _ := command.Flags().GetString("format")
	if err := printer.ValidatePrintTableFormat(format); err != nil {
		return err
	}

	eventType, _ := command.Flags().GetString("type")
	limit, _ := command.Flags().GetInt("limit")
	sinceStr, _ := command.Flags().GetString("since")
	untilStr, _ := command.Flags().GetString("until")

	var sinceTime, untilTime time.Time
	if sinceStr != "" {
		sinceTime, err = parseTimeArg(sinceStr)
		if err != nil {
			return fmt.Errorf("--since: %w", err)
		}
	}
	if untilStr != "" {
		untilTime, err = parseTimeArg(untilStr)
		if err != nil {
			return fmt.Errorf("--until: %w", err)
		}
	}

	var componentID int
	if len(args) > 0 {
		comp, err := client.GetComponentByUUID(args[0])
		if err != nil {
			return fmt.Errorf("component %s: %w", args[0], err)
		}
		if comp.ID == nil {
			return errors.New("component has no ID")
		}
		componentID = *comp.ID
	} else {
		comp, err := pickComponentByPage(client, statusPage.ID)
		if err != nil {
			return err
		}
		if comp.ID == nil {
			return errors.New("component has no ID")
		}
		componentID = *comp.ID
	}

	apiEntries, err := client.GetComponentLogEntries(componentID, statusPage.ID, eventType)
	if err != nil {
		return err
	}

	var entries []printer.LogEntry
	for _, e := range apiEntries {
		if !inPeriod(e.At, sinceTime, untilTime) {
			continue
		}
		entries = append(entries, printer.LogEntry{
			At:       e.At,
			UUID:     e.UUID,
			Object:   e.Object,
			Title:    e.Title,
			Status:   e.Status,
			Desc:     e.Desc,
			ParentID: e.ParentID,
		})
	}

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	cfg := printer.NewPrintTableConfig()
	cfg.Format = format
	cfg.PrintBlockTotal = false

	return printer.PrintLogs(os.Stdout, entries, cfg)
}

func pickComponentByPage(client *api.Client, statusPageID int) (*api.Component, error) {
	components, err := client.GetPaginatedComponents(
		api.NewAllPaginatedRequest(api.PaginatedRequestFilter{"status_page": statusPageID}),
	)
	if err != nil {
		return nil, err
	}
	if len(components.Results) == 0 {
		return nil, errors.New("no components found for this page")
	}

	items := make([]string, len(components.Results))
	for i, c := range components.Results {
		uuid := derefStr(c.UUID)
		items[i] = fmt.Sprintf("%-40s  %s  (%s)", c.Name, string(c.Impact), uuid)
	}

	prompt := promptui.Select{
		Label: "Select component",
		Items: items,
		Size:  10,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &components.Results[idx], nil
}

func inPeriod(t, since, until time.Time) bool {
	if !since.IsZero() && t.Before(since) {
		return false
	}
	if !until.IsZero() && t.After(until) {
		return false
	}
	return true
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func parseTimeArg(s string) (time.Time, error) {
	// Try absolute formats first so "2026-05-20" isn't misread as natural language
	for _, f := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	t, err := naturaldate.Parse(s, time.Now(), naturaldate.WithDirection(naturaldate.Past))
	if err != nil {
		return time.Time{}, fmt.Errorf("unrecognized format %q (try: today, yesterday, last week, 2 days ago, 3 months ago, or 2026-05-20)", s)
	}
	return t, nil
}
