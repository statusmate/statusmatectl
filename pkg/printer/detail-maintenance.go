package printer

import (
	"fmt"
	"io"
	"strings"

	"github.com/statusmate/statusmatectl/pkg/api"
)

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
