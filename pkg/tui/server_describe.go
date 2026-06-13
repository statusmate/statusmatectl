package tui

import (
	"strconv"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// ServerDescribeView shows the stored authRC of a single configured server.
type ServerDescribeView struct {
	app    *App
	detail *tview.Table
}

func newServerDescribeView(app *App) *ServerDescribeView {
	d := &ServerDescribeView{app: app}

	d.detail = tview.NewTable().SetSelectable(true, false)
	d.detail.SetBorder(true)
	d.detail.SetTitle(" Server AuthRC ")
	d.detail.SetTitleAlign(tview.AlignCenter)
	d.detail.SetBackgroundColor(tcell.ColorBlack)
	d.detail.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			app.popPage()
			app.tv.SetFocus(app.servers.table)
			return nil
		}
		return ev
	})
	app.pages.AddPage(viewServerDescribe, d.detail, true, false)

	return d
}

func (d *ServerDescribeView) show(s *serverEntry) {
	d.detail.Clear()

	row := 0
	set := func(label, value string) {
		d.detail.SetCell(row, 0, detailLabelCell(label))
		d.detail.SetCell(row, 1, detailValueCell(value))
		row++
	}
	orDash := func(v string) string {
		if v == "" {
			return "-"
		}
		return v
	}

	set("Server", s.domain)
	if s.authRC == nil {
		set("Auth", "not configured")
	} else {
		set("API", orDash(s.authRC.API))
		set("Token", orDash(s.authRC.Token))
		set("Default Status Page", orDash(s.authRC.DefaultStatusPage))
		set("Default Release Page", orDash(s.authRC.DefaultReleasePage))
		set("Default Team", strconv.Itoa(s.authRC.DefaultTeam))
		d.renderRecentPages(&row, s.authRC.RecentPages)
	}

	d.app.pushPage(viewServerDescribe)
	d.app.tv.SetFocus(d.detail)
}

// renderRecentPages appends a "Recent Pages" sub-table listing recently visited
// status-page slugs, most-recent first, starting at *row.
func (d *ServerDescribeView) renderRecentPages(row *int, recent []string) {
	if len(recent) == 0 {
		return
	}

	*row++ // blank separator row
	header := tview.NewTableCell("RECENT PAGES").
		SetTextColor(tcell.ColorYellow).
		SetAttributes(tcell.AttrBold).
		SetSelectable(false)
	d.detail.SetCell(*row, 0, header)
	*row++

	for i := len(recent) - 1; i >= 0; i-- {
		d.detail.SetCell(*row, 0, detailValueCell(recent[i]))
		*row++
	}
}
