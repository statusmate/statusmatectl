package printer

import (
	"fmt"
	"io"
	"strings"

	"github.com/statusmate/statusmatectl/pkg/api"
)

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
