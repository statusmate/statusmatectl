package tui

import (
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
	tv           *tview.Application
	pages        *tview.Pages
	client       *api.Client
	statusPage   *api.StatusPage
	version      string
	srvInfo      *ServerInfo
	pageSwitcher *PageSwitcher
	navTabs      *NavTabs
	help         *Help
	current      string
	incidents    *IncidentsView
	components   *ComponentsView
	maintenance  *MaintenanceView
	team         *TeamView
	user         *api.User
}

// NewApp creates and initializes the TUI application.
func NewApp(client *api.Client, statusPage *api.StatusPage, version string) *App {
	a := &App{
		tv:         tview.NewApplication(),
		pages:      tview.NewPages(),
		client:     client,
		statusPage: statusPage,
		version:    version,
	}
	a.build()
	return a
}

func (a *App) build() {
	a.srvInfo = newServerInfo(a)
	a.pageSwitcher = newPageSwitcher(a)
	a.navTabs = newNavTabs(a)
	a.help = newHelp(a)

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
		AddItem(a.buildHeader(), 5, 0, false).
		AddItem(a.pages, 0, 1, true)

	a.tv.SetRoot(layout, true)
	a.tv.SetInputCapture(a.onGlobalKey)
	a.switchTo(viewIncidents)

	go func() {
		user, err := a.client.GetMe()
		if err == nil {
			a.user = user
			a.tv.QueueUpdateDraw(a.renderHeader)
		}
	}()

	go func() {
		result, err := a.client.GetPaginatedStatusPages(api.NewAllPaginatedRequest(nil))
		if err != nil || len(result.Results) == 0 {
			return
		}
		a.tv.QueueUpdateDraw(func() {
			a.pageSwitcher.setPages(result.Results)
		})
	}()
}

func (a *App) buildHeader() tview.Primitive {
	header := tview.NewFlex()
	header.SetDirection(tview.FlexColumn)
	header.AddItem(a.srvInfo, clusterInfoWidth, 1, false)
	header.AddItem(a.pageSwitcher, pageSwitcherWidth, 1, false)
	header.AddItem(a.navTabs, navTabsWidth, 1, false)
	header.AddItem(a.help, 0, 1, false)
	return header
}

func (a *App) renderHeader() {
	a.srvInfo.render()
	a.pageSwitcher.render()
	a.navTabs.render()
	a.help.render()
}

func (a *App) switchStatusPage(idx int) {
	pages := a.pageSwitcher.pages
	if idx < 0 || idx >= len(pages) {
		return
	}
	a.statusPage = &pages[idx]
	a.renderHeader()
	a.refreshCurrent()
}

func (a *App) onGlobalKey(ev *tcell.EventKey) *tcell.EventKey {
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

	if r := ev.Rune(); r >= '0' && r <= '9' {
		a.switchStatusPage(int(r - '0'))
		return nil
	}

	return ev
}

func (a *App) switchTo(name string) {
	a.current = name
	a.pages.SwitchToPage(name)
	a.renderHeader()
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
