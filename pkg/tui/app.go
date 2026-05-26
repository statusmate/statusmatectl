package tui

import (
	"fmt"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

const (
	viewIncidents   = "incidents"
	viewComponents  = "components"
	viewMaintenance = "maintenance"
	viewTeam        = "team"
)

// App is the main TUI application.
type App struct {
	tv          *tview.Application
	pages       *tview.Pages
	client      *api.Client
	statusPage  *api.StatusPage
	header      *tview.TextView
	footer      *tview.TextView
	current     string
	incidents   *IncidentsView
	components  *ComponentsView
	maintenance *MaintenanceView
	team        *TeamView
}

// NewApp creates and initializes the TUI application.
func NewApp(client *api.Client, statusPage *api.StatusPage) *App {
	a := &App{
		tv:         tview.NewApplication(),
		pages:      tview.NewPages(),
		client:     client,
		statusPage: statusPage,
	}
	a.build()
	return a
}

func (a *App) build() {
	a.header = tview.NewTextView().SetDynamicColors(true)
	a.footer = tview.NewTextView().SetDynamicColors(true)

	a.incidents = newIncidentsView(a)
	a.components = newComponentsView(a)
	a.maintenance = newMaintenanceView(a)
	a.team = newTeamView(a)

	a.pages.AddPage(viewIncidents, a.incidents.root(), true, true)
	a.pages.AddPage(viewComponents, a.components.root(), true, false)
	a.pages.AddPage(viewMaintenance, a.maintenance.root(), true, false)
	a.pages.AddPage(viewTeam, a.team.root(), true, false)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.header, 1, 0, false).
		AddItem(a.pages, 0, 1, true).
		AddItem(a.footer, 1, 0, false)

	a.tv.SetRoot(layout, true)
	a.tv.SetInputCapture(a.onGlobalKey)
	a.switchTo(viewIncidents)
}

func (a *App) onGlobalKey(ev *tcell.EventKey) *tcell.EventKey {
	// Only intercept global keys when on a main view (not inside a modal)
	name, _ := a.pages.GetFrontPage()
	switch name {
	case viewIncidents, viewComponents, viewMaintenance, viewTeam:
	default:
		return ev
	}

	switch ev.Rune() {
	case 'q':
		a.tv.Stop()
		return nil
	case 'i':
		a.switchTo(viewIncidents)
		return nil
	case 'c':
		a.switchTo(viewComponents)
		return nil
	case 'm':
		a.switchTo(viewMaintenance)
		return nil
	case 't':
		a.switchTo(viewTeam)
		return nil
	case 'r':
		a.refreshCurrent()
		return nil
	}
	return ev
}

func (a *App) switchTo(name string) {
	a.current = name
	a.pages.SwitchToPage(name)
	a.renderHeader()
	a.renderFooter()
	a.refreshCurrent()
}

func (a *App) refreshCurrent() {
	switch a.current {
	case viewIncidents:
		a.incidents.refresh()
	case viewComponents:
		a.components.refresh()
	case viewMaintenance:
		a.maintenance.refresh()
	case viewTeam:
		a.team.refresh()
	}
}

func (a *App) renderHeader() {
	spName := "-"
	if a.statusPage != nil {
		spName = a.statusPage.Name
	}

	navItems := []struct{ key, label string }{
		{viewIncidents, "Incidents"},
		{viewComponents, "Components"},
		{viewMaintenance, "Maintenance"},
		{viewTeam, "Team"},
	}

	text := fmt.Sprintf("[yellow::b]st4[-:-:-]  [aqua]%s[-]  ", spName)
	for _, item := range navItems {
		if item.key == a.current {
			text += fmt.Sprintf("[black:yellow] %s [-:-]  ", item.label)
		} else {
			text += fmt.Sprintf("[gray]%s[-]  ", item.label)
		}
	}
	a.header.SetText(text)
}

var footerHelp = map[string]string{
	viewIncidents:   "[yellow]<enter>[-] detail  [yellow]n[-] new incident  [yellow]u[-] add update  [yellow]i/c/m/t[-] switch  [yellow]r[-] refresh  [yellow]q[-] quit",
	viewComponents:  "[yellow]<enter>[-] detail  [yellow]i/c/m/t[-] switch  [yellow]r[-] refresh  [yellow]q[-] quit",
	viewMaintenance: "[yellow]<enter>[-] detail  [yellow]i/c/m/t[-] switch  [yellow]r[-] refresh  [yellow]q[-] quit",
	viewTeam:        "[yellow]i/c/m/t[-] switch  [yellow]r[-] refresh  [yellow]q[-] quit",
}

func (a *App) renderFooter() {
	a.footer.SetText(footerHelp[a.current])
}

func (a *App) showModal(name string, p tview.Primitive, width, height int) {
	centered := centeredBox(p, width, height)
	a.pages.AddPage(name, centered, true, true)
	a.tv.SetFocus(p)
}

func (a *App) closeModal(name string) {
	a.pages.RemovePage(name)
	a.tv.SetFocus(a.pages)
}

func centeredBox(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 0, true).
			AddItem(nil, 0, 1, false),
			width, 0, true).
		AddItem(nil, 0, 1, false)
}

// Run starts the TUI event loop.
func (a *App) Run() error {
	return a.tv.Run()
}
