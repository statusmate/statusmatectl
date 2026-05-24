package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/printer"
)

var ShowUpdateCmd = &cobra.Command{
	Use:   "show-update <uuid>",
	Short: "Show details of a single update and its parent incident or maintenance",
	Args:  cobra.ExactArgs(1),
	RunE:  showUpdateCmdF,
}

func init() {
	RootCmd.AddCommand(ShowUpdateCmd)
}

func showUpdateCmdF(command *cobra.Command, args []string) error {
	client, err := InitClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	uuid := args[0]

	update, err := client.GetUpdateByUUID(uuid)
	if err != nil {
		return err
	}

	w := os.Stdout
	color := printer.IsTerminal(w)

	// Print update header
	shortID := update.UUID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	header := "update " + shortID
	if color {
		fmt.Fprintf(w, "\033[33m%s\033[0m\n", header)
	} else {
		fmt.Fprintln(w, header)
	}

	fmt.Fprintf(w, "UUID:      %s\n", update.UUID)
	fmt.Fprintf(w, "Date:      %s\n", update.At.Local().Format("Mon, 02 Jan 2006 15:04:05"))
	fmt.Fprintf(w, "Status:    %s\n", update.Status)
	if update.Description != "" {
		desc := strings.ReplaceAll(strings.TrimSpace(update.Description), "\n", " ")
		fmt.Fprintf(w, "Message:   %s\n", desc)
	}
	if len(update.Components) > 0 {
		fmt.Fprintln(w, "Components:")
		for _, c := range update.Components {
			uuid := ""
			if c.UUID != nil {
				uuid = *c.UUID
			}
			fmt.Fprintf(w, "  - component=%d uuid=%s impact=%s\n", c.Component, uuid, c.Impact)
		}
	}

	fmt.Fprintln(w)

	// Print parent entity
	switch {
	case update.Incident != nil:
		inc, err := client.GetIncidentByID(*update.Incident)
		if err != nil {
			fmt.Fprintf(w, "incident id=%d (details unavailable: %v)\n", *update.Incident, err)
			return nil
		}
		if color {
			fmt.Fprintf(w, "\033[36mincident\033[0m\n")
		} else {
			fmt.Fprintln(w, "incident")
		}
		if inc.UUID != nil {
			fmt.Fprintf(w, "UUID:      %s\n", *inc.UUID)
		}
		fmt.Fprintf(w, "Title:     %s\n", inc.Title)
		fmt.Fprintf(w, "Status:    %s\n", inc.Status)
		fmt.Fprintf(w, "StartAt:   %s\n", inc.StartAt.Local().Format("Mon, 02 Jan 2006 15:04:05"))
		if inc.EndAt != nil {
			fmt.Fprintf(w, "EndAt:     %s\n", inc.EndAt.Local().Format("Mon, 02 Jan 2006 15:04:05"))
		}
		if inc.AbsoluteURL != nil {
			fmt.Fprintf(w, "URL:       %s\n", *inc.AbsoluteURL)
		}

	case update.Maintenance != nil:
		m, err := client.GetMaintenanceByID(*update.Maintenance)
		if err != nil {
			fmt.Fprintf(w, "maintenance id=%d (details unavailable: %v)\n", *update.Maintenance, err)
			return nil
		}
		if color {
			fmt.Fprintf(w, "\033[36mmaintenance\033[0m\n")
		} else {
			fmt.Fprintln(w, "maintenance")
		}
		if m.UUID != nil {
			fmt.Fprintf(w, "UUID:      %s\n", *m.UUID)
		}
		fmt.Fprintf(w, "Title:     %s\n", m.Title)
		fmt.Fprintf(w, "Status:    %s\n", m.Status)
		if m.StartAt != nil {
			fmt.Fprintf(w, "StartAt:   %s\n", m.StartAt.Local().Format("Mon, 02 Jan 2006 15:04:05"))
		}
		if m.EndAt != nil {
			fmt.Fprintf(w, "EndAt:     %s\n", m.EndAt.Local().Format("Mon, 02 Jan 2006 15:04:05"))
		}
		fmt.Fprintf(w, "URL:       %s\n", m.AbsoluteURL)
	}

	return nil
}
