package tui

import (
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

type TemplateDescribeView struct {
	app  *App
	text *tview.TextView
	flex *tview.Flex
}

func newTemplateDescribeView(app *App) *TemplateDescribeView {
	d := &TemplateDescribeView{app: app}

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
	d.flex.SetTitle(" Template Describe ")
	d.flex.SetTitleAlign(tview.AlignCenter)
	d.flex.SetBackgroundColor(tcell.ColorBlack)

	app.pages.AddPage(viewTmplDescribe, d.flex, true, false)
	return d
}

func (d *TemplateDescribeView) show(t *api.Template) {
	d.text.SetText("[#808080::]Loading...[-::]")
	d.app.pushPage(viewTmplDescribe)
	d.app.tv.SetFocus(d.text)

	uuid := ""
	if t.UUID != nil {
		uuid = *t.UUID
	}

	go func() {
		var fresh *api.Template
		if uuid != "" {
			if f, err := d.app.client.GetTemplate(uuid); err == nil {
				fresh = f
			}
		}
		if fresh == nil {
			fresh = t
		}

		var comps []api.Component
		var spFilter api.PaginatedRequestFilter
		if d.app.statusPage != nil {
			spFilter = api.PaginatedRequestFilter{"status_page": d.app.statusPage.ID}
		}
		cached := make([]api.Component, len(d.app.components.components))
		copy(cached, d.app.components.components)
		if len(cached) > 0 {
			comps = cached
		} else if len(fresh.Components) > 0 {
			filter := api.PaginatedRequestFilter{}
			if spFilter != nil {
				filter = spFilter
			}
			if res, err := d.app.client.GetPaginatedComponents(api.NewAllPaginatedRequest(filter)); err == nil {
				comps = res.Results
			}
		}

		d.app.tv.QueueUpdateDraw(func() {
			d.render(fresh, comps)
		})
	}()
}

func (d *TemplateDescribeView) render(t *api.Template, comps []api.Component) {
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

	field("Title", tview.Escape(t.Title))
	if t.UUID != nil {
		field("UUID", tview.Escape(*t.UUID))
	}
	if t.FriendlyName != "" {
		field("Friendly_name", tview.Escape(t.FriendlyName))
	}
	if t.Status != nil && *t.Status != "" {
		fieldColor("Status", prettyStatus(*t.Status), colorTag(templateStatusColor(*t.Status)))
	}
	v, c := boolVal(t.Notify)
	fieldColor("Notify", v, c)
	if t.CreatedAt != nil {
		field("Created_at", t.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	if t.UpdatedAt != nil {
		field("Updated_at", t.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	if t.Description != "" {
		sb.WriteString("\n")
		fmt.Fprintf(&sb, "[yellow::b]Description:[-::-]\n")
		fmt.Fprintf(&sb, "[white::]%s[-::]\n", tview.Escape(t.Description))
	}

	if len(t.Components) > 0 {
		nameMap := make(map[int]string, len(comps))
		for _, c := range comps {
			if c.ID != nil {
				nameMap[*c.ID] = c.Name
			}
		}
		sb.WriteString("\n")
		fmt.Fprintf(&sb, "[yellow::b]Components:[-::-]\n")
		for _, ac := range t.Components {
			name := fmt.Sprintf("id:%d", ac.Component)
			if n, ok := nameMap[ac.Component]; ok {
				name = n
			}
			color := colorTag(impactColor(ac.Impact))
			fmt.Fprintf(&sb, "  [white::]%-30s[-::]  [%s::]%s[-::]\n",
				tview.Escape(name), color, string(ac.Impact))
		}
	}

	if len(t.AssignedTags) > 0 {
		sb.WriteString("\n")
		fmt.Fprintf(&sb, "[yellow::b]Tags:[-::-]\n")
		for _, tag := range t.AssignedTags {
			fmt.Fprintf(&sb, "  [white::]%s[-::]\n", tview.Escape(tag.Title))
		}
	}

	d.text.SetText(sb.String())
	d.text.ScrollToBeginning()
}
