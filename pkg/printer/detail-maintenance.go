package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/statusmate/statusmatectl/pkg/api"
)

func PrintDetailMaintenance(w io.Writer, m *api.Maintenance, format string) error {
	if format == PrintTableFormatJSON {
		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON: %w", err)
		}
		_, err = fmt.Fprintln(w, string(data))
		return err
	}

	color := IsTerminal(w)

	shortID := ""
	if m.UUID != nil {
		shortID = *m.UUID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
	}

	header := "maintenance " + shortID
	if color {
		fmt.Fprintf(w, "\033[36m%s\033[0m\n", header)
	} else {
		fmt.Fprintln(w, header)
	}

	if m.UUID != nil {
		fmt.Fprintf(w, "UUID:         %s\n", *m.UUID)
	}
	fmt.Fprintf(w, "Title:        %s\n", m.Title)
	fmt.Fprintf(w, "Status:       %s\n", m.Status)
	if m.StartAt != nil {
		fmt.Fprintf(w, "StartAt:      %s\n", m.StartAt.Local().Format("Mon, 02 Jan 2006 15:04:05"))
	}
	if m.EndAt != nil {
		fmt.Fprintf(w, "EndAt:        %s\n", m.EndAt.Local().Format("Mon, 02 Jan 2006 15:04:05"))
	}
	if m.Description != "" {
		fmt.Fprintf(w, "Description:  %s\n", strings.TrimSpace(m.Description))
	}
	fmt.Fprintf(w, "URL:          %s\n", m.AbsoluteURL)
	fmt.Fprintf(w, "Notify:       %v\n", m.Notify)
	fmt.Fprintf(w, "AutoStart:    %v\n", m.AutoStart)
	fmt.Fprintf(w, "AutoEnd:      %v\n", m.AutoEnd)
	fmt.Fprintf(w, "AffectUptime: %v\n", m.AffectUptime)

	if len(m.Components) > 0 {
		fmt.Fprintln(w, "Components:")
		for _, c := range m.Components {
			uuid := ""
			if c.UUID != nil {
				uuid = " uuid=" + *c.UUID
			}
			fmt.Fprintf(w, "  - component=%d%s impact=%s\n", c.Component, uuid, c.Impact)
		}
	}

	if len(m.Updates) > 0 {
		fmt.Fprintln(w)
		if color {
			fmt.Fprintf(w, "\033[33mUpdates (%d)\033[0m\n", len(m.Updates))
		} else {
			fmt.Fprintf(w, "Updates (%d)\n", len(m.Updates))
		}
		for _, u := range m.Updates {
			fmt.Fprintln(w)
			shortUUID := u.UUID
			if len(shortUUID) > 8 {
				shortUUID = shortUUID[:8]
			}
			fmt.Fprintf(w, "  update %s\n", shortUUID)
			fmt.Fprintf(w, "  Date:    %s\n", u.At.Local().Format("Mon, 02 Jan 2006 15:04:05"))
			fmt.Fprintf(w, "  Status:  %s\n", u.Status)
			if u.Description != "" {
				desc := strings.ReplaceAll(strings.TrimSpace(u.Description), "\n", " ")
				fmt.Fprintf(w, "  Message: %s\n", desc)
			}
		}
	}

	return nil
}

func PrintSummaryMaintenance(w io.Writer, m *api.Maintenance) error {
	components := make([]string, 0, len(m.Components))
	for _, c := range m.Components {
		components = append(components, fmt.Sprintf("%s:%d", c.Impact, c.Component))
	}

	summary := fmt.Sprintf(
		"uuid=%s\n"+
			"url=%s\n"+
			"title=%s\n"+
			"components=%s\n"+
			"status=%s\n"+
			"notify=%v\n"+
			"auto_start=%v\n"+
			"auto_end=%v\n"+
			"affect_uptime=%v\n"+
			"start_at=%s\n"+
			"created_at=%s",
		nullOrValue(m.UUID),
		m.AbsoluteURL,
		m.Title,
		strings.Join(components, ", "),
		string(m.Status),
		m.Notify,
		m.AutoStart,
		m.AutoEnd,
		m.AffectUptime,
		formatTime(m.StartAt),
		formatTime(m.CreatedAt),
	)

	_, err := fmt.Fprintln(w, summary)
	return err
}

func PrintSummaryCreateMaintenancePayload(w io.Writer, m *api.CreateMaintenancePayload) error {
	summary := fmt.Sprintf(
		"title=%s\n"+
			"description=%s\n"+
			"components=%s\n"+
			"start_at=%s\n"+
			"end_at=%s\n"+
			"notify=%v\n"+
			"auto_start=%v\n"+
			"auto_end=%v\n"+
			"affect_uptime=%v",
		m.Title,
		m.Description,
		strings.Join(m.Components, ", "),
		m.StartAt,
		m.EndAt,
		m.Notify,
		m.AutoStart,
		m.AutoEnd,
		m.AffectUptime,
	)

	_, err := fmt.Fprintln(w, summary)
	return err
}
