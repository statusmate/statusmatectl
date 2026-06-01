package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

func openInBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start() //nolint:errcheck
}

func shortUUID(uuid string) string {
	if len(uuid) > 8 {
		return uuid[:8]
	}
	return uuid
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}

	r := []rune(s)

	if len(r) <= max {
		return s
	}

	if max <= 3 {
		return string(r[:max])
	}

	return string(r[:max-3]) + "..."
}

func formatAge(t *time.Time) string {
	if t == nil {
		return "-"
	}
	d := time.Since(*t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format("2006-01-02 15:04")
}

func prettyStatus(s string) string {
	s = strings.TrimPrefix(s, "incident_")
	s = strings.TrimPrefix(s, "maintenance_")
	s = strings.ReplaceAll(s, "_", " ")
	if len(s) > 0 {
		s = strings.ToUpper(s[:1]) + s[1:]
	}
	return s
}

func formatIncidentStatus(s api.IncidentStatusType) string {
	return prettyStatus(string(s))
}

func formatMaintenanceStatus(s api.MaintenanceStatusType) string {
	return prettyStatus(string(s))
}

func incidentStatusColor(s api.IncidentStatusType) tcell.Color {
	switch s {
	case api.IncidentStatusResolved:
		return tcell.ColorGreen
	case api.IncidentStatusInvestigation:
		return tcell.ColorRed
	case api.IncidentStatusMonitoring:
		return tcell.ColorYellow
	case api.IncidentStatusIdentified:
		return tcell.ColorOrange
	default:
		return tcell.ColorWhite
	}
}

func maintenanceStatusColor(s api.MaintenanceStatusType) tcell.Color {
	switch s {
	case api.MaintenanceStatusCompleted:
		return tcell.ColorGreen
	case api.MaintenanceStatusInProgress:
		return tcell.ColorYellow
	case api.MaintenanceStatusNotStarted:
		return tcell.ColorGray
	default:
		return tcell.ColorWhite
	}
}

func detailLabelCell(t string) *tview.TableCell {
	return tview.NewTableCell(t + ":").
		SetTextColor(tcell.ColorBlue).
		SetAlign(tview.AlignLeft)
}

func detailValueCell(t string) *tview.TableCell {
	return tview.NewTableCell(t).
		SetTextColor(tcell.ColorWhite).
		SetExpansion(1)
}

func detailSectionCell(t string) *tview.TableCell {
	return tview.NewTableCell(t).
		SetTextColor(tcell.ColorYellow).
		SetAttributes(tcell.AttrBold).
		SetExpansion(2)
}

func colorTag(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func impactColor(s api.ImpactType) tcell.Color {
	switch s {
	case api.ImpactTypeOperational:
		return tcell.ColorGreen
	case api.ImpactTypeMajorOutage:
		return tcell.ColorRed
	case api.ImpactTypePartialOutage:
		return tcell.ColorOrange
	case api.ImpactTypeDegradedPerformance:
		return tcell.ColorYellow
	case api.ImpactTypeUnderMaintenance:
		return tcell.ColorBlue
	default:
		return tcell.ColorWhite
	}
}
