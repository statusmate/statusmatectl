package tui

import (
	"strings"

	"github.com/derailed/tview"
)

// BreadcrumbsView renders the navigation stack as a one-line bar at the bottom of the TUI.
type BreadcrumbsView struct {
	*tview.TextView
	app *App
}

func newBreadcrumbsView(app *App) *BreadcrumbsView {
	tv := tview.NewTextView().SetDynamicColors(true)
	tv.SetBorderPadding(0, 0, 1, 0)
	return &BreadcrumbsView{TextView: tv, app: app}
}

// StackPushed implements StackListener.
func (b *BreadcrumbsView) StackPushed(_ string) {}

// StackPopped implements StackListener.
func (b *BreadcrumbsView) StackPopped(_ string) {}

// StackTop implements StackListener — re-renders after any navigation change.
func (b *BreadcrumbsView) StackTop(_ string) {
	b.render()
}

func (b *BreadcrumbsView) render() {
	stack := b.app.pages.Peek()
	parts := make([]string, len(stack))
	for i, name := range stack {
		if i == len(stack)-1 {
			parts[i] = "[yellow::b]" + pageDisplayName(name) + "[-:-:-]"
		} else {
			parts[i] = "[gray::]" + pageDisplayName(name) + "[-:-:-]"
		}
	}
	b.SetText(" " + strings.Join(parts, " [gray::]>[-:-:-] "))
}

func pageDisplayName(name string) string {
	switch name {
	case viewIncidents:
		return "Incidents"
	case viewComponents:
		return "Components"
	case viewMaintenance:
		return "Maintenance"
	case viewTeam:
		return "Team"
	case viewServers:
		return "Servers"
	case requestViewLogs:
		return "Logs"
	case "incDetail":
		return "Incident"
	case "compDetail":
		return "Component"
	case "maintDetail":
		return "Maintenance"
	case "logDetail":
		return "Log Entry"
	default:
		return name
	}
}
