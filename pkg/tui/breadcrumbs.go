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

	var parts []string
	for i, name := range stack {
		label := pageDisplayName(name)

		if i == len(stack)-1 {
			parts = append(parts,
				"[black:yellow:b] <"+label+"> [-:-:-]",
			)
		} else {
			parts = append(parts,
				"[black:navy:b] <"+label+"> [-:-:-]",
			)
		}
	}

	b.SetText(strings.Join(parts, " "))
}

func pageDisplayName(name string) string {
	switch name {
	case viewIncidents:
		return "Incidents"
	case viewComponents:
		return "Components"
	case viewMaintenance:
		return "Maintenances"
	case viewTeam:
		return "Teams"
	case viewServers:
		return "Servers"
	case viewSubscribers:
		return "Subscribers"
	case viewRequestLogs:
		return "Logs"
	case "incDetail":
		return "Incident"
	case "compDetail":
		return "Component"
	case "maintDetail":
		return "Maintenance"
	case "logDetail":
		return "Log Entry"
	case viewTmplDescribe:
		return "Template"
	case viewPubPageDescribe:
		return "Public Page"
	case viewServerDescribe:
		return "Server AuthRC"
	case viewSubDescribe:
		return "Subscriber"
	default:
		return name
	}
}
