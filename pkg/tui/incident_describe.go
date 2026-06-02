package tui

import (
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

type IncidentDescribeView struct {
	app  *App
	text *tview.TextView
	flex *tview.Flex
}

func newIncidentDescribeView(app *App) *IncidentDescribeView {
	d := &IncidentDescribeView{app: app}

	d.text = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false).
		SetWordWrap(true)

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
	d.flex.SetTitle(" Incident Describe ")
	d.flex.SetTitleAlign(tview.AlignCenter)
	d.flex.SetBackgroundColor(tcell.ColorBlack)

	app.pages.AddPage(viewIncDescribe, d.flex, true, false)
	return d
}

func (d *IncidentDescribeView) show(inc *api.Incident) {
	d.text.SetText("[#808080::]Loading...[-::]")
	d.app.pushPage(viewIncDescribe)
	d.app.tv.SetFocus(d.text)

	uuid := ""
	if inc.UUID != nil {
		uuid = *inc.UUID
	}
	incID := 0
	if inc.ID != nil {
		incID = *inc.ID
	}

	go func() {
		var fresh *api.Incident
		if uuid != "" {
			if f, err := d.app.client.GetIncidentByUUID(uuid); err == nil {
				fresh = f
			}
		}
		if fresh == nil {
			fresh = inc
		}

		var updates []api.Update[string]
		if incID != 0 {
			if res, err := d.app.client.GetPaginatedUpdates(
				api.NewAllPaginatedRequest(api.PaginatedRequestFilter{"incident": incID}),
			); err == nil {
				updates = res.Results
			}
		}

		d.app.tv.QueueUpdateDraw(func() {
			d.render(fresh, updates)
		})
	}()
}

func (d *IncidentDescribeView) render(inc *api.Incident, updates []api.Update[string]) {
	const keyWidth = 16
	var sb strings.Builder

	field := func(key, val string) {
		fmt.Fprintf(&sb, "[yellow::b]%-*s[-::-]  [white::]%s[-::]\n", keyWidth, key+":", val)
	}
	fieldColor := func(key, val, color string) {
		fmt.Fprintf(&sb, "[yellow::b]%-*s[-::-]  [%s::]%s[-::]\n", keyWidth, key+":", color, val)
	}
	boolVal := func(b bool) (string, string) {
		if b {
			return "true", "green"
		}
		return "false", "#808080"
	}

	field("Title", tview.Escape(inc.Title))
	fieldColor("Status", formatIncidentStatus(inc.Status), colorTag(incidentStatusColor(inc.Status)))
	if inc.UUID != nil {
		field("UUID", tview.Escape(*inc.UUID))
	}
	if inc.CreatedAt != nil {
		field("Created_at", inc.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	if !inc.StartAt.IsZero() {
		field("Start_at", inc.StartAt.Format("2006-01-02 15:04:05"))
	}
	if inc.EndAt != nil {
		field("End_at", inc.EndAt.Format("2006-01-02 15:04:05"))
	}
	if inc.Description != "" {
		field("Description", tview.Escape(inc.Description))
	}
	if inc.PrivateNote != "" {
		field("Private_note", tview.Escape(inc.PrivateNote))
	}
	v, c := boolVal(inc.Notify)
	fieldColor("Notify", v, c)
	v, c = boolVal(inc.ShowOnTop)
	fieldColor("Show_on_top", v, c)
	v, c = boolVal(inc.AffectUptime)
	fieldColor("Affect_uptime", v, c)

	if len(updates) == 0 {
		d.text.SetText(sb.String())
		return
	}

	const (
		statusW = 16
		timeW   = 18
	)

	sb.WriteString("\n")
	fmt.Fprintf(&sb, "[yellow::b]Updates: [-::-]\n")
	fmt.Fprintf(&sb, "[#808080::]%-*s  %-*s  %s[-::]\n", statusW, "STATUS", timeW, "TIME", "DESCRIPTION")
	fmt.Fprintf(&sb, "[#808080::]%s  %s  %s[-::]\n",
		strings.Repeat("─", statusW),
		strings.Repeat("─", timeW),
		strings.Repeat("─", 50),
	)
	for _, u := range updates {
		status := api.IncidentStatusType(u.Status)
		color := colorTag(incidentStatusColor(status))
		fmt.Fprintf(&sb, "[%s::]%-*s[-::]  [white::]%-*s[-::]  [white::]%s[-::]\n",
			color, statusW, formatIncidentStatus(status),
			timeW, u.At.Format("2006-01-02 15:04"),
			tview.Escape(u.Description),
		)
	}

	d.text.SetText(sb.String())
	d.text.ScrollToBeginning()
}
