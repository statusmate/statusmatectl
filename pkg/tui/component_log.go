package tui

import (
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

const viewCompLog = "compLog"

type ComponentLogView struct {
	app  *App
	text *tview.TextView
	flex *tview.Flex
}

func newComponentLogView(app *App) *ComponentLogView {
	d := &ComponentLogView{app: app}

	d.text = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false)

	d.text.SetBackgroundColor(tcell.ColorBlack)
	d.text.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			app.popPage()
			return nil
		}
		return ev
	})

	d.flex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(d.text, 0, 1, true)
	d.flex.SetBorder(true)
	d.flex.SetTitle(" Component Log ")
	d.flex.SetTitleAlign(tview.AlignCenter)
	d.flex.SetBackgroundColor(tcell.ColorBlack)

	app.pages.AddPage(viewCompLog, d.flex, true, false)
	return d
}

func (d *ComponentLogView) show(comp *api.Component) {
	compName := comp.Name
	d.flex.SetTitle(fmt.Sprintf(" Log: %s ", tview.Escape(compName)))
	d.text.SetText("[#808080::]Loading...[-::]")
	d.app.pushPage(viewCompLog)
	d.app.tv.SetFocus(d.text)

	if comp.ID == nil || d.app.statusPage == nil {
		d.text.SetText("[red::]No component ID or status page[-::]")
		return
	}
	compID := *comp.ID
	statusPageID := d.app.statusPage.ID

	go func() {
		entries, err := d.app.client.GetComponentLogEntries(compID, statusPageID, "")
		d.app.tv.QueueUpdateDraw(func() {
			if err != nil {
				d.text.SetText(fmt.Sprintf("[red::]Error: %s[-::]", tview.Escape(err.Error())))
				return
			}
			d.render(compName, entries)
		})
	}()
}

func (d *ComponentLogView) render(compName string, entries []api.ComponentLogEntry) {
	if len(entries) == 0 {
		d.text.SetText(fmt.Sprintf("[#808080::]No log entries for %s[-::]", tview.Escape(compName)))
		d.text.ScrollToBeginning()
		return
	}

	type group struct {
		object  string
		title   string
		entries []api.ComponentLogEntry
	}

	var order []int
	groups := map[int]*group{}
	for _, e := range entries {
		if _, ok := groups[e.ParentID]; !ok {
			order = append(order, e.ParentID)
			groups[e.ParentID] = &group{object: e.Object, title: e.Title}
		}
		groups[e.ParentID].entries = append(groups[e.ParentID].entries, e)
	}

	const (
		statusW = 14
		timeW   = 16
	)

	var sb strings.Builder

	for i, id := range order {
		g := groups[id]

		if i > 0 {
			sb.WriteString("\n")
		}

		objectColor := logObjectColor(g.object)
		fmt.Fprintf(&sb, "[%s::b][%s][-::-] [white::b]%s[-::-]\n",
			objectColor, strings.ToUpper(g.object), tview.Escape(g.title))

		fmt.Fprintf(&sb, "[#808080::]  %-*s  %-*s  %s[-::]\n",
			statusW, "STATUS", timeW, "TIME", "DESCRIPTION")
		fmt.Fprintf(&sb, "[#808080::]  %s  %s  %s[-::]\n",
			strings.Repeat("─", statusW),
			strings.Repeat("─", timeW),
			strings.Repeat("─", 50),
		)

		for _, e := range g.entries {
			status := prettyStatus(e.Status)
			statusColor := logStatusColor(e.Object, e.Status)
			desc := strings.ReplaceAll(strings.TrimSpace(e.Desc), "\n", " ")
			desc = truncate(desc, 80)

			fmt.Fprintf(&sb, "  [%s::]%-*s[-::]  [#aaaaaa::]%-*s[-::]  [white::]%s[-::]\n",
				colorTag(statusColor), statusW, status,
				timeW, e.At.Local().Format("2006-01-02 15:04"),
				tview.Escape(desc),
			)
		}
	}

	d.text.SetText(sb.String())
	d.text.ScrollToBeginning()
}

func logObjectColor(object string) string {
	switch object {
	case "incident":
		return colorTag(tcell.ColorRed)
	case "maintenance":
		return colorTag(tcell.ColorLightBlue)
	default:
		return colorTag(tcell.ColorYellow)
	}
}

func logStatusColor(object, status string) tcell.Color {
	if object == "incident" {
		return incidentStatusColor(api.IncidentStatusType(status))
	}
	return maintenanceStatusColor(api.MaintenanceStatusType(status))
}
