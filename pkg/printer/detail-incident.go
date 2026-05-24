package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/statusmate/statusmatectl/pkg/api"
)

func PrintDetailIncident(w io.Writer, incident *api.Incident, format string) error {
	if format == PrintTableFormatJSON {
		data, err := json.MarshalIndent(incident, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON: %w", err)
		}
		_, err = fmt.Fprintln(w, string(data))
		return err
	}

	color := IsTerminal(w)

	shortID := ""
	if incident.UUID != nil {
		shortID = *incident.UUID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
	}

	header := "incident " + shortID
	if color {
		fmt.Fprintf(w, "\033[36m%s\033[0m\n", header)
	} else {
		fmt.Fprintln(w, header)
	}

	if incident.UUID != nil {
		fmt.Fprintf(w, "UUID:         %s\n", *incident.UUID)
	}
	fmt.Fprintf(w, "Title:        %s\n", incident.Title)
	fmt.Fprintf(w, "Status:       %s\n", incident.Status)
	fmt.Fprintf(w, "StartAt:      %s\n", incident.StartAt.Local().Format("Mon, 02 Jan 2006 15:04:05"))
	if incident.EndAt != nil {
		fmt.Fprintf(w, "EndAt:        %s\n", incident.EndAt.Local().Format("Mon, 02 Jan 2006 15:04:05"))
	}
	if incident.Description != "" {
		fmt.Fprintf(w, "Description:  %s\n", strings.TrimSpace(incident.Description))
	}
	if incident.AbsoluteURL != nil {
		fmt.Fprintf(w, "URL:          %s\n", *incident.AbsoluteURL)
	}
	fmt.Fprintf(w, "Notify:       %v\n", incident.Notify)
	fmt.Fprintf(w, "ShowOnTop:    %v\n", incident.ShowOnTop)
	fmt.Fprintf(w, "AffectUptime: %v\n", incident.AffectUptime)

	if len(incident.Components) > 0 {
		fmt.Fprintln(w, "Components:")
		for _, c := range incident.Components {
			uuid := ""
			if c.UUID != nil {
				uuid = " uuid=" + *c.UUID
			}
			fmt.Fprintf(w, "  - component=%d%s impact=%s\n", c.Component, uuid, c.Impact)
		}
	}

	if len(incident.Updates) > 0 {
		fmt.Fprintln(w)
		if color {
			fmt.Fprintf(w, "\033[33mUpdates (%d)\033[0m\n", len(incident.Updates))
		} else {
			fmt.Fprintf(w, "Updates (%d)\n", len(incident.Updates))
		}
		for _, u := range incident.Updates {
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

func PrintSummaryIncident(w io.Writer, incident *api.Incident) error {
	components := make([]string, 0, len(incident.Components))
	for _, c := range incident.Components {
		components = append(components, fmt.Sprintf("%s:%d", c.Impact, c.Component))
	}

	summary := fmt.Sprintf(
		"uuid=%s\n"+
			"url=%s\n"+
			"title=%s\n"+
			"components=%s\n"+
			"status=%s\n"+
			"notify=%v\n"+
			"show_on_top=%v\n"+
			"affect_uptime=%v\n"+
			"start_at=%s\n"+
			"created_at=%s",
		nullOrValue(incident.UUID),
		nullOrValue(incident.AbsoluteURL),
		incident.Title,
		strings.Join(components, ", "),
		string(incident.Status),
		incident.Notify,
		incident.ShowOnTop,
		incident.AffectUptime,
		formatTime(&incident.StartAt),
		formatTime(incident.CreatedAt),
	)

	_, err := fmt.Fprintln(w, summary)
	return err
}

func PrintSummaryCreateIncidentPayload(w io.Writer, incident *api.CreateIncidentPayload) error {
	summary := fmt.Sprintf(
		"title=%s\n"+
			"description=%s\n"+
			"components=%s\n"+
			"status=%s\n"+
			"notify=%v\n"+
			"show_on_top=%v\n"+
			"affect_uptime=%v\n"+
			"start_at=%s",
		incident.Title,
		incident.Description,
		strings.Join(incident.Components, ", "),
		incident.Status,
		incident.Notify,
		incident.ShowOnTop,
		incident.AffectUptime,
		formatTime(&incident.StartAt),
	)

	_, err := fmt.Fprintln(w, summary)
	return err
}
