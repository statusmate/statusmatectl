package tui

import (
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

type MaintenanceDescribeView struct {
	app  *App
	text *tview.TextView
	flex *tview.Flex
}

func newMaintenanceDescribeView(app *App) *MaintenanceDescribeView {
	d := &MaintenanceDescribeView{app: app}

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
	d.flex.SetTitle(" Maintenance Describe ")
	d.flex.SetTitleAlign(tview.AlignCenter)
	d.flex.SetBackgroundColor(tcell.ColorBlack)

	app.pages.AddPage(viewMaintDescribe, d.flex, true, false)
	return d
}

func (d *MaintenanceDescribeView) show(m *api.Maintenance) {
	d.text.SetText("[#808080::]Loading...[-::]")
	d.app.pushPage(viewMaintDescribe)
	d.app.tv.SetFocus(d.text)

	uuid := ""
	if m.UUID != nil {
		uuid = *m.UUID
	}
	maintID := 0
	if m.ID != nil {
		maintID = *m.ID
	}

	go func() {
		var fresh *api.Maintenance
		if uuid != "" {
			if f, err := d.app.client.GetMaintenanceByUUID(uuid); err == nil {
				fresh = f
			}
		}
		if fresh == nil {
			fresh = m
		}

		var updates []api.Update[string]
		if maintID != 0 {
			if res, err := d.app.client.GetPaginatedUpdates(
				api.NewAllPaginatedRequest(api.PaginatedRequestFilter{"maintenance": maintID}),
			); err == nil {
				updates = res.Results
			}
		}

		d.app.tv.QueueUpdateDraw(func() {
			d.render(fresh, updates)
		})
	}()
}

func (d *MaintenanceDescribeView) render(m *api.Maintenance, updates []api.Update[string]) {
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

	field("Title", tview.Escape(m.Title))
	fieldColor("Status", formatMaintenanceStatus(m.Status), colorTag(maintenanceStatusColor(m.Status)))
	if m.UUID != nil {
		field("UUID", tview.Escape(*m.UUID))
	}
	if m.CreatedAt != nil {
		field("Created_at", m.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	if m.StartAt != nil {
		field("Start_at", m.StartAt.Format("2006-01-02 15:04:05"))
	}
	if m.EndAt != nil {
		field("End_at", m.EndAt.Format("2006-01-02 15:04:05"))
	}
	if m.Description != "" {
		field("Description", tview.Escape(m.Description))
	}
	if m.PrivateNote != "" {
		field("Private_note", tview.Escape(m.PrivateNote))
	}
	v, c := boolVal(m.Notify)
	fieldColor("Notify", v, c)
	v, c = boolVal(m.AutoStart)
	fieldColor("Auto_start", v, c)
	v, c = boolVal(m.AutoEnd)
	fieldColor("Auto_end", v, c)
	v, c = boolVal(m.AffectUptime)
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
		status := api.MaintenanceStatusType(u.Status)
		color := colorTag(maintenanceStatusColor(status))
		fmt.Fprintf(&sb, "[%s::]%-*s[-::]  [white::]%-*s[-::]  [white::]%s[-::]\n",
			color, statusW, formatMaintenanceStatus(status),
			timeW, u.At.Format("2006-01-02 15:04"),
			tview.Escape(u.Description),
		)
	}

	d.text.SetText(sb.String())
	d.text.ScrollToBeginning()
}
