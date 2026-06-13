package tui

import (
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

// ServersView displays all locally configured servers with their auth info.
type ServersView struct {
	app        *App
	table      *tview.Table
	describe   *ServerDescribeView
	servers    []serverEntry
	displayed  []serverEntry
	filterText string
}

type serverEntry struct {
	domain  string
	authRC  *api.AuthRC
	current bool
}

func newServersView(app *App) *ServersView {
	v := &ServersView{app: app}
	v.table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSelectedStyle(tcell.StyleDefault.
			Background(tcell.ColorNavy).
			Foreground(tcell.ColorWhite))
	v.table.SetBorder(true)
	v.table.SetTitle(" Servers ")
	v.table.SetTitleAlign(tview.AlignCenter)
	v.table.SetInputCapture(v.onKey)
	v.table.SetBackgroundColor(tcell.ColorBlack)

	v.describe = newServerDescribeView(app)

	return v
}

func (v *ServersView) root() tview.Primitive { return v.table }

func (v *ServersView) refresh() {
	go func() {
		domains, err := listAvailableServers()
		if err != nil {
			return
		}
		currentDomain := sanitizeDomain(strings.TrimPrefix(
			strings.TrimPrefix(v.app.client.BaseURL, "https://"), "http://"))

		entries := make([]serverEntry, 0, len(domains))
		for _, d := range domains {
			rc, _ := loadServerAuthRC(d)
			entries = append(entries, serverEntry{
				domain:  d,
				authRC:  rc,
				current: d == currentDomain,
			})
		}
		v.servers = entries
		v.app.tv.QueueUpdateDraw(v.render)
	}()
}

func (v *ServersView) filter(text string) {
	v.filterText = text
	v.render()
}

func (v *ServersView) clearFilter() {
	v.filterText = ""
	v.render()
}

func (v *ServersView) render() {
	lower := strings.ToLower(v.filterText)
	v.displayed = v.displayed[:0]
	for _, s := range v.servers {
		if lower == "" || strings.Contains(strings.ToLower(s.domain), lower) {
			v.displayed = append(v.displayed, s)
		}
	}

	if lower != "" {
		v.table.SetTitle(fmt.Sprintf(" Servers [%d/%d] ", len(v.displayed), len(v.servers)))
	} else {
		v.table.SetTitle(fmt.Sprintf(" Servers [%d] ", len(v.servers)))
	}
	v.table.Clear()

	for i, h := range []string{"SERVER", "TOKEN", "DEFAULT PAGE", "STATUS"} {
		v.table.SetCell(0, i, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false).
			SetExpansion(1))
	}

	for i, s := range v.displayed {
		row := i + 1
		token := "-"
		defaultPage := "-"
		if s.authRC != nil {
			if t := s.authRC.Token; t != "" {
				if len(t) > 8 {
					token = "****" + t[len(t)-4:]
				} else {
					token = "****"
				}
			}
			if s.authRC.DefaultStatusPage != "" {
				defaultPage = s.authRC.DefaultStatusPage
			}
		}

		domainColor := tcell.ColorWhite
		statusText := ""
		statusColor := tcell.ColorGray
		if s.current {
			domainColor = tcell.ColorYellow
			statusText = "● active"
			statusColor = tcell.ColorGreen
		}

		v.table.SetCell(row, 0, tview.NewTableCell(s.domain).SetTextColor(domainColor).SetExpansion(3))
		v.table.SetCell(row, 1, tview.NewTableCell(token).SetTextColor(tcell.ColorGray).SetExpansion(2))
		v.table.SetCell(row, 2, tview.NewTableCell(defaultPage).SetTextColor(tcell.ColorAqua).SetExpansion(2))
		v.table.SetCell(row, 3, tview.NewTableCell(statusText).SetTextColor(statusColor))
	}

	if len(v.displayed) > 0 {
		v.table.Select(1, 0)
	}
}

func (v *ServersView) selected() *serverEntry {
	row, _ := v.table.GetSelection()
	if row <= 0 || row-1 >= len(v.displayed) {
		return nil
	}
	return &v.displayed[row-1]
}

func (v *ServersView) onKey(ev *tcell.EventKey) *tcell.EventKey {
	if ev.Key() == tcell.KeyEnter {
		if s := v.selected(); s != nil && !s.current {
			v.app.switchServer(s.domain)
		}
		return nil
	}
	if (ev.Key() == 'l') {
		if s := v.selected(); s != nil && !s.current {
			v.app.switchServer(s.domain)
			v.app.switchTo(viewRequestLogs);
		}
		return nil
	}
	if ev.Key() == tcell.KeyRune && ev.Rune() == 'd' {
		if s := v.selected(); s != nil {
			v.describe.show(s)
		}
		return nil
	}
	return ev
}
