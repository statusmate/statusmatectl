package tui

import (
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

type ComponentDescribeView struct {
	app  *App
	text *tview.TextView
	flex *tview.Flex
}

func newComponentDescribeView(app *App) *ComponentDescribeView {
	d := &ComponentDescribeView{app: app}

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
	d.flex.SetTitle(" Component Describe ")
	d.flex.SetTitleAlign(tview.AlignCenter)
	d.flex.SetBackgroundColor(tcell.ColorBlack)

	app.pages.AddPage(viewCompDescribe, d.flex, true, false)
	return d
}

func (d *ComponentDescribeView) show(comp *api.Component) {
	d.text.SetText("[#808080::]Loading...[-::]")
	d.app.pushPage(viewCompDescribe)
	d.app.tv.SetFocus(d.text)

	uuid := ""
	if comp.UUID != nil {
		uuid = *comp.UUID
	}

	go func() {
		var fresh *api.Component
		if uuid != "" {
			if f, err := d.app.client.GetComponentByUUID(uuid); err == nil {
				fresh = f
			}
		}
		if fresh == nil {
			fresh = comp
		}

		d.app.tv.QueueUpdateDraw(func() {
			d.render(fresh)
		})
	}()
}

func (d *ComponentDescribeView) render(comp *api.Component) {
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

	field("Name", tview.Escape(comp.Name))
	if comp.UUID != nil {
		field("UUID", tview.Escape(*comp.UUID))
	}
	fieldColor("Impact", string(comp.Impact), colorTag(impactColor(comp.Impact)))
	v, c := boolVal(comp.Enabled)
	fieldColor("Enabled", v, c)
	if comp.Uptime != "" {
		field("Uptime", comp.Uptime)
	}
	v, c = boolVal(comp.Histogram)
	fieldColor("Histogram", v, c)
	v, c = boolVal(comp.Collapse)
	fieldColor("Collapse", v, c)
	v, c = boolVal(comp.Private)
	fieldColor("Private", v, c)
	if comp.StartDate != nil && *comp.StartDate != "" {
		field("Start_date", *comp.StartDate)
	}
	if comp.CreatedAt != nil {
		field("Created_at", comp.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	if comp.UpdatedAt != nil {
		field("Updated_at", comp.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	if comp.Description != "" {
		sb.WriteString("\n")
		fmt.Fprintf(&sb, "[yellow::b]Description:[-::-]\n")
		fmt.Fprintf(&sb, "[white::]%s[-::]\n", tview.Escape(comp.Description))
	}

	d.text.SetText(sb.String())
	d.text.ScrollToBeginning()
}
