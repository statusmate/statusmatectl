package tui

import (
	"fmt"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

// ComponentsView displays a list of components.
type ComponentsView struct {
	app        *App
	table      *tview.Table
	components []api.Component
}

func newComponentsView(app *App) *ComponentsView {
	v := &ComponentsView{app: app}

	v.table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSelectedStyle(tcell.StyleDefault.
			Background(tcell.ColorNavy).
			Foreground(tcell.ColorWhite))

	v.table.SetInputCapture(v.onKey)
	return v
}

func (v *ComponentsView) root() tview.Primitive { return v.table }

func (v *ComponentsView) refresh() {
	go func() {
		filter := api.PaginatedRequestFilter{}
		if v.app.statusPage != nil {
			filter["status_page"] = v.app.statusPage.ID
		}
		result, err := v.app.client.GetPaginatedComponents(api.NewAllPaginatedRequest(filter))
		if err != nil {
			return
		}
		v.components = result.Results
		v.app.tv.QueueUpdateDraw(v.render)
	}()
}

func (v *ComponentsView) render() {
	v.table.Clear()

	for i, h := range []string{"NAME", "IMPACT", "ENABLED", "UPTIME", "UPDATED"} {
		v.table.SetCell(0, i, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false).
			SetExpansion(1))
	}

	for i, comp := range v.components {
		row := i + 1
		enabled := "yes"
		if !comp.Enabled {
			enabled = "no"
		}
		uptime := comp.Uptime
		if uptime == "" {
			uptime = "-"
		}
		v.table.SetCell(row, 0, tview.NewTableCell(comp.Name).SetTextColor(tcell.ColorWhite).SetExpansion(3))
		v.table.SetCell(row, 1, tview.NewTableCell(string(comp.Impact)).SetTextColor(impactColor(comp.Impact)).SetExpansion(2))
		v.table.SetCell(row, 2, tview.NewTableCell(enabled).SetTextColor(tcell.ColorGray))
		v.table.SetCell(row, 3, tview.NewTableCell(uptime).SetTextColor(tcell.ColorGray).SetExpansion(1))
		v.table.SetCell(row, 4, tview.NewTableCell(formatAge(comp.UpdatedAt)).SetTextColor(tcell.ColorGray))
	}

	if len(v.components) > 0 {
		v.table.Select(1, 0)
	}
}

func (v *ComponentsView) selected() *api.Component {
	row, _ := v.table.GetSelection()
	if row <= 0 || row-1 >= len(v.components) {
		return nil
	}
	return &v.components[row-1]
}

func (v *ComponentsView) onKey(ev *tcell.EventKey) *tcell.EventKey {
	if ev.Key() == tcell.KeyEnter {
		if comp := v.selected(); comp != nil {
			v.showDetail(comp)
		}
		return nil
	}
	return ev
}

func (v *ComponentsView) showDetail(comp *api.Component) {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)

	uuid := "-"
	if comp.UUID != nil {
		uuid = *comp.UUID
	}

	text := fmt.Sprintf("[yellow::b]%s[-:-:-]\n\n", comp.Name)
	text += fmt.Sprintf("[blue]UUID:[-]    %s\n", uuid)
	text += fmt.Sprintf("[blue]Impact:[-]  %s\n", comp.Impact)
	text += fmt.Sprintf("[blue]Enabled:[-] %v\n", comp.Enabled)
	text += fmt.Sprintf("[blue]Uptime:[-]  %s\n", comp.Uptime)
	if comp.Description != "" {
		text += fmt.Sprintf("\n[blue]Description:[-]\n%s\n", comp.Description)
	}

	tv.SetText(text)
	tv.SetBorder(true).SetTitle(" Component Detail ").SetTitleAlign(tview.AlignLeft)
	tv.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			v.app.closeModal("compDetail")
		}
		return ev
	})
	v.app.showModal("compDetail", tv, 70, 20)
}
