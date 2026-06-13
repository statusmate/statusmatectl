package tui

import (
	"fmt"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

// SubscriberDescribeView shows the detail of a single subscriber.
type SubscriberDescribeView struct {
	app    *App
	detail *tview.Table
}

func newSubscriberDescribeView(app *App) *SubscriberDescribeView {
	d := &SubscriberDescribeView{app: app}

	d.detail = tview.NewTable().SetSelectable(true, false)
	d.detail.SetBorder(true)
	d.detail.SetTitle(" Subscriber Detail ")
	d.detail.SetTitleAlign(tview.AlignCenter)
	d.detail.SetBackgroundColor(tcell.ColorBlack)
	d.detail.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			app.popPage()
			app.tv.SetFocus(app.subscribers.table)
			return nil
		}
		return ev
	})
	app.pages.AddPage(viewSubDescribe, d.detail, true, false)

	return d
}

func (d *SubscriberDescribeView) show(s *api.Subscriber) {
	d.detail.Clear()

	row := 0
	set := func(label, value string) {
		d.detail.SetCell(row, 0, detailLabelCell(label))
		d.detail.SetCell(row, 1, detailValueCell(value))
		row++
	}

	set("UUID", nullOrDash(s.UUID))
	set("Email", s.Email)
	set("Subscribe by email", yesNo(s.SubscribeByEmail))
	set("Subscribe by webhook", yesNo(s.SubscribeByWebhook))
	if s.WebhookURL != "" {
		set("Webhook URL", s.WebhookURL)
	}
	set("Has password", yesNo(s.HasPassword))

	d.detail.SetCell(row, 0, detailLabelCell("Verified"))
	d.detail.SetCell(row, 1, tview.NewTableCell(yesNo(s.Confirmed)).
		SetTextColor(verifiedColor(s.Confirmed)).SetExpansion(1))
	row++

	if s.SubscribeTo != "" {
		set("Subscribe to", s.SubscribeTo)
	}
	set("Created", formatTimePtr(s.CreatedAt))

	d.app.pushPage(viewSubDescribe)
	d.app.tv.SetFocus(d.detail)

	if s.SubscribeTo == "selected" && len(s.Components) > 0 {
		d.loadComponents(s.Components, row)
	}
}

// loadComponents resolves the subscribed component IDs to names and appends a
// Components section to the detail table, starting at startRow.
func (d *SubscriberDescribeView) loadComponents(ids []int, startRow int) {
	var filter api.PaginatedRequestFilter
	if d.app.statusPage != nil {
		filter = api.PaginatedRequestFilter{"status_page": d.app.statusPage.ID}
	}

	go func() {
		comps := make([]api.Component, len(d.app.components.components))
		copy(comps, d.app.components.components)
		if len(comps) == 0 {
			if res, err := d.app.client.GetPaginatedComponents(api.NewAllPaginatedRequest(filter)); err == nil {
				comps = res.Results
			}
		}

		nameMap := make(map[int]string, len(comps))
		for _, c := range comps {
			if c.ID != nil {
				nameMap[*c.ID] = c.Name
			}
		}

		d.app.tv.QueueUpdateDraw(func() {
			row := startRow
			row++ // blank spacer row
			d.detail.SetCell(row, 0, detailLabelCell("Components"))
			row++
			for _, id := range ids {
				name := fmt.Sprintf("id:%d", id)
				if n, ok := nameMap[id]; ok {
					name = n
				}
				d.detail.SetCell(row, 1, detailValueCell(tview.Escape(name)))
				row++
			}
		})
	}()
}
