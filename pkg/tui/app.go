package tui

import (
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

const (
	viewIncidents    = "incidents"
	viewComponents   = "components"
	viewMaintenance  = "maintenance"
	viewTeam         = "team"
	viewServers      = "servers"
	viewTemplates    = "templates"
	viewPublicPages  = "public-pages"
	requestViewLogs  = "request-logs"
)

// App is the main TUI application.
type App struct {
	tv           *tview.Application
	pages        *Pages
	layout       *tview.Flex
	client       *api.Client
	statusPage   *api.StatusPage
	version      string
	srvInfo      *ServerInfo
	pageSwitcher *PageSwitcher
	navTabs      *NavTabs
	pageActions  *PageActions
	header       *tview.Flex
	current      string
	incidents    *IncidentsView
	components   *ComponentsView
	maintenance  *MaintenanceView
	team         *TeamView
	servers      *ServersView
	templates    *TemplatesView
	publicPages  *PublicPagesView
	logs         *RequestLogView
	user         *api.User
	prompt       *CommandPrompt
	breadcrumbs  *BreadcrumbsView
}

// NewApp creates and initializes the TUI application.
func NewApp(client *api.Client, statusPage *api.StatusPage, version string) *App {
	a := &App{
		tv:         tview.NewApplication(),
		pages:      newPages(),
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
	a.pageActions = newPageActions(a)
	a.breadcrumbs = newBreadcrumbsView(a)

	a.incidents = newIncidentsView(a)
	a.components = newComponentsView(a)
	a.maintenance = newMaintenanceView(a)
	a.team = newTeamView(a)
	a.servers = newServersView(a)
	a.templates = newTemplatesView(a)
	a.publicPages = newPublicPagesView(a)
	a.logs = newRequestLogView(a)
	a.prompt = newCommandPrompt(a)

	a.pages.AddPage(viewIncidents, a.incidents.root(), true, true)
	a.pages.AddPage(viewComponents, a.components.root(), true, false)
	a.pages.AddPage(viewMaintenance, a.maintenance.root(), true, false)
	a.pages.AddPage(viewTeam, a.team.root(), true, false)
	a.pages.AddPage(viewServers, a.servers.root(), true, false)
	a.pages.AddPage(viewTemplates, a.templates.root(), true, false)
	a.pages.AddPage(viewPublicPages, a.publicPages.root(), true, false)
	a.pages.AddPage(requestViewLogs, a.logs.root(), true, false)

	a.pages.addListener(a.breadcrumbs)

	a.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.buildHeader(), 5, 0, false).
		AddItem(a.prompt, 0, 0, false).
		AddItem(a.pages, 0, 1, true).
		AddItem(a.breadcrumbs, 1, 0, false)

	a.tv.SetRoot(a.layout, true)
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
	a.header = tview.NewFlex()
	a.header.SetDirection(tview.FlexColumn)
	a.header.AddItem(a.srvInfo, clusterInfoWidth, 1, false)
	a.header.AddItem(a.pageSwitcher, pageSwitcherWidth, 1, false)
	a.header.AddItem(a.navTabs, navTabsWidth, 1, false)
	a.header.AddItem(a.pageActions, 0, 1, false)
	return a.header
}

func (a *App) renderHeader() {
	a.srvInfo.render()
	if a.current == requestViewLogs {
		a.header.ResizeItem(a.pageSwitcher, 0, 0)
		a.header.ResizeItem(a.navTabs, 0, 0)
	} else {
		a.header.ResizeItem(a.pageSwitcher, pageSwitcherWidth, 1)
		a.header.ResizeItem(a.navTabs, navTabsWidth, 1)
		a.pageSwitcher.render()
		a.navTabs.render()
	}
	a.pageActions.render()
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
	if a.prompt.active {
		return ev
	}

	name, _ := a.pages.GetFrontPage()
	switch name {
	case viewIncidents, viewComponents, viewMaintenance, viewTeam, viewServers, viewTemplates, viewPublicPages, requestViewLogs:
	default:
		return ev
	}

	switch ev.Rune() {
	case ':':
		a.prompt.ActivateCommand()
		return nil
	case '/':
		f, c := a.currentSearch()
		a.prompt.ActivateSearch(f, c)
		return nil
	case 'q':
		a.logs.stopTailing()
		a.tv.Stop()
		return nil
	case 'i':
		a.switchTo(viewIncidents)
		return nil
	case 't':
		a.switchTo(viewTemplates)
		return nil
	case 'c':
		a.switchTo(viewComponents)
		return nil
	case 'm':
		a.switchTo(viewMaintenance)
		return nil
	case 'r':
		a.refreshCurrent()
		return nil
	}

	if ev = a.pageActions.handleKey(ev); ev == nil {
		return nil
	}

	if r := ev.Rune(); r >= '0' && r <= '9' {
		a.switchStatusPage(int(r - '0'))
		return nil
	}

	return ev
}

func (a *App) currentSearch() (func(string), func()) {
	switch a.current {
	case viewIncidents:
		return a.incidents.filter, a.incidents.clearFilter
	case viewComponents:
		return a.components.filter, a.components.clearFilter
	case viewMaintenance:
		return a.maintenance.filter, a.maintenance.clearFilter
	case viewTeam:
		return a.team.filter, a.team.clearFilter
	case viewServers:
		return a.servers.filter, a.servers.clearFilter
	case viewTemplates:
		return a.templates.filter, a.templates.clearFilter
	case viewPublicPages:
		return a.publicPages.filter, a.publicPages.clearFilter
	case requestViewLogs:
		return a.logs.filter, a.logs.clearFilter
	}
	return func(string) {}, func() {}
}

func (a *App) switchTo(name string) {
	if a.current == requestViewLogs && name != requestViewLogs {
		a.logs.stopTailing()
	}
	a.current = name
	a.pages.Reset(name)
	a.renderHeader()
	a.refreshCurrent()
}

func (a *App) pushPage(name string) {
	a.pages.Push(name)
}

func (a *App) popPage() {
	a.pages.Pop()
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
	case viewServers:
		a.servers.refresh()
	case viewTemplates:
		a.templates.refresh()
	case viewPublicPages:
		a.publicPages.refresh()
	case requestViewLogs:
		a.logs.refresh()
	}
}

func (a *App) switchServer(domain string) {
	newClient, err := loadServerClient(domain, a.client)
	if err != nil {
		return
	}
	a.client = newClient
	a.user = nil
	a.statusPage = nil
	a.pageSwitcher.pages = nil
	a.renderHeader()
	a.refreshCurrent()

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
			if a.statusPage == nil && len(result.Results) > 0 {
				a.statusPage = &result.Results[0]
				a.renderHeader()
				a.refreshCurrent()
			}
		})
	}()
}


func (a *App) Quit() {
	a.logs.stopTailing()
	a.tv.Stop()
}

// Run starts the TUI event loop.
func (a *App) Run() error {
	return a.tv.Run()
}
