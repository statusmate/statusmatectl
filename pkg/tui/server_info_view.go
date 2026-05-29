package tui

import (
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

const (
	clusterInfoWidth  = 36
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
