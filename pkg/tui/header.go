package tui

import (
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

const (
	clusterInfoWidth  = 36
	pageSwitcherWidth = 26
	navTabsWidth      = 20
)

// ServerInfo is the left panel of the header showing server/user/page.
type ServerInfo struct {
	*tview.Table
	app *App
}

func newServerInfo(app *App) *ServerInfo {
	s := &ServerInfo{
		Table: tview.NewTable().SetSelectable(false, false),
		app:   app,
	}
	s.SetBorderPadding(0, 0, 1, 0)
	return s
}

func (s *ServerInfo) render() {
	a := s.app

	server := strings.TrimPrefix(a.client.BaseURL, "https://")
	server = strings.TrimPrefix(server, "http://")

	spName := "-"
	if a.statusPage != nil {
		spName = a.statusPage.Name
	}

	userName := "-"
	if a.user != nil {
		if a.user.Email != "" {
			userName = a.user.Email
		} else if a.user.Username != "" {
			userName = a.user.Username
		}
	}

	ver := a.version
	if ver == "" {
		ver = "dev"
	}

	rows := [][2]string{
		{"Server", server},
		{"User", userName},
		{"Page", spName},
		{"Version", ver},
	}

	s.Clear()
	for i, r := range rows {
		s.SetCell(i, 0, s.labelCell(r[0]))
		s.SetCell(i, 1, s.valueCell(r[1]))
	}
}

func (s *ServerInfo) labelCell(t string) *tview.TableCell {
	return tview.NewTableCell(t+":").
		SetAlign(tview.AlignLeft).
		SetTextColor(tcell.ColorYellow)
}

func (s *ServerInfo) valueCell(t string) *tview.TableCell {
	return tview.NewTableCell(t).
		SetExpansion(1).
		SetTextColor(tcell.ColorAqua)
}

// PageSwitcher shows numbered status pages (max 5) for quick navigation via digit keys.
type PageSwitcher struct {
	*tview.Table
	app   *App
	pages []api.StatusPage
}

func newPageSwitcher(app *App) *PageSwitcher {
	p := &PageSwitcher{
		Table: tview.NewTable().SetSelectable(false, false),
		app:   app,
	}
	p.SetBorderPadding(0, 0, 1, 0)
	return p
}

func (p *PageSwitcher) setPages(pages []api.StatusPage) {
	p.pages = pages
	p.render()
}

func (p *PageSwitcher) render() {
	p.Clear()
	limit := len(p.pages)
	if limit > 5 {
		limit = 5
	}
	for i := 0; i < limit; i++ {
		pg := p.pages[i]

		keyCell := tview.NewTableCell(fmt.Sprintf("<%d>", i)).
			SetTextColor(tcell.ColorMediumPurple)

		nameCell := tview.NewTableCell(" " + pg.Name).SetExpansion(1)
		if p.app.statusPage != nil && p.app.statusPage.ID == pg.ID {
			nameCell.SetTextColor(tcell.ColorYellow).SetAttributes(tcell.AttrBold)
		} else {
			nameCell.SetTextColor(tcell.ColorWhite)
		}

		p.SetCell(i, 0, keyCell)
		p.SetCell(i, 1, nameCell)
	}
}

// NavTabs is the navigation panel showing view shortcuts in two columns.
type NavTabs struct {
	*tview.Table
	app *App
}

func newNavTabs(app *App) *NavTabs {
	n := &NavTabs{
		Table: tview.NewTable().SetSelectable(false, false),
		app:   app,
	}
	n.SetBorderPadding(0, 0, 1, 0)
	return n
}

func (n *NavTabs) render() {
	items := []struct{ view, label, key string }{
		{viewIncidents, "Incidents", "i"},
		{viewComponents, "Components", "c"},
		{viewMaintenance, "Maintenance", "m"},
	}

	n.Clear()
	for row, item := range items {
		keyCell := tview.NewTableCell(fmt.Sprintf("<%s>", item.key)).
			SetTextColor(tcell.ColorCornflowerBlue)

		nameCell := tview.NewTableCell(" " + item.label).SetExpansion(1)
		if n.app.current == item.view {
			nameCell.SetTextColor(tcell.ColorYellow).SetAttributes(tcell.AttrBold)
		} else {
			nameCell.SetTextColor(tcell.ColorWhite)
		}

		n.SetCell(row, 0, keyCell)
		n.SetCell(row, 1, nameCell)
	}
}

// Help shows global and view-specific shortcuts as a two-column table.
type Help struct {
	*tview.Table
	app *App
}

func newHelp(app *App) *Help {
	h := &Help{
		Table: tview.NewTable().SetSelectable(false, false),
		app:   app,
	}
	h.SetBorderPadding(0, 0, 1, 0)
	return h
}

func (h *Help) render() {
	a := h.app

	add := func(row int, key, action string) {
		h.SetCell(row, 0, tview.NewTableCell(fmt.Sprintf("<%s>", key)).
			SetTextColor(tcell.ColorYellow))
		h.SetCell(row, 1, tview.NewTableCell(" "+action).
			SetExpansion(1).
			SetTextColor(tcell.ColorWhite))
	}

	h.Clear()
	add(0, "r", "Refresh")
	add(1, "q", "Quit")

	viewKeys := map[string][][2]string{
		viewIncidents:   {{"enter", "Detail"}, {"n", "New"}, {"u", "Update"}},
		viewComponents:  {{"enter", "Detail"}},
		viewMaintenance: {{"enter", "Detail"}},
		viewTeam:        {},
	}
	for i, kv := range viewKeys[a.current] {
		add(2+i, kv[0], kv[1])
	}
}
