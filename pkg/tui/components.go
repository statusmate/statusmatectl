package tui

import (
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

// ComponentsView displays a list of components.
type ComponentsView struct {
	app          *App
	table        *tview.Table
	detail       *tview.Table
	components   []api.Component
	displayed    []api.Component
	filterText   string
	pendingNavID *int
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
	v.table.SetBorder(true)
	v.table.SetTitle(" Components ")
	v.table.SetTitleAlign(tview.AlignCenter)
	v.table.SetInputCapture(v.onKey)

	v.detail = tview.NewTable().SetSelectable(true, false)
	v.detail.SetBorder(true)
	v.detail.SetTitle(" Component Detail ")
	v.detail.SetTitleAlign(tview.AlignCenter)
	v.detail.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			app.pages.SwitchToPage(viewComponents)
			app.tv.SetFocus(v.table)
		}
		return ev
	})
	app.pages.AddPage("compDetail", v.detail, true, false)

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

func (v *ComponentsView) filter(text string) {
	v.filterText = text
	v.render()
}

func (v *ComponentsView) clearFilter() {
	v.filterText = ""
	v.render()
}

func (v *ComponentsView) render() {
	lower := strings.ToLower(v.filterText)
	v.displayed = v.displayed[:0]
	for _, comp := range v.components {
		if lower == "" ||
			strings.Contains(strings.ToLower(comp.Name), lower) ||
			strings.Contains(strings.ToLower(string(comp.Impact)), lower) {
			v.displayed = append(v.displayed, comp)
		}
	}

	if lower != "" {
		v.table.SetTitle(fmt.Sprintf(" Components [%d/%d] ", len(v.displayed), len(v.components)))
	} else {
		v.table.SetTitle(fmt.Sprintf(" Components [%d] ", len(v.components)))
	}
	v.table.Clear()

	for i, h := range []string{"NAME", "IMPACT", "ENABLED", "UPTIME", "UPDATED"} {
		v.table.SetCell(0, i, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false).
			SetExpansion(1))
	}

	for i, comp := range v.displayed {
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

	if len(v.displayed) > 0 {
		v.table.Select(1, 0)
	}
	v.doNavigate()
}

func (v *ComponentsView) navigateTo(id int) {
	for i, comp := range v.displayed {
		if comp.ID != nil && *comp.ID == id {
			v.table.Select(i+1, 0)
			return
		}
	}
	v.pendingNavID = &id
}

func (v *ComponentsView) doNavigate() {
	if v.pendingNavID == nil {
		return
	}
	for i, comp := range v.displayed {
		if comp.ID != nil && *comp.ID == *v.pendingNavID {
			v.table.Select(i+1, 0)
			v.pendingNavID = nil
			return
		}
	}
}

func (v *ComponentsView) selected() *api.Component {
	row, _ := v.table.GetSelection()
	if row <= 0 || row-1 >= len(v.displayed) {
		return nil
	}
	return &v.displayed[row-1]
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
	v.detail.Clear()

	uuid := "-"
	if comp.UUID != nil {
		uuid = *comp.UUID
	}
	enabled := "yes"
	if !comp.Enabled {
		enabled = "no"
	}
	uptime := comp.Uptime
	if uptime == "" {
		uptime = "-"
	}

	row := 0
	v.detail.SetCell(row, 0, detailSectionCell(comp.Name))
	row++
	v.detail.SetCell(row, 0, detailLabelCell("UUID"))
	v.detail.SetCell(row, 1, detailValueCell(uuid))
	row++
	v.detail.SetCell(row, 0, detailLabelCell("Impact"))
	v.detail.SetCell(row, 1, tview.NewTableCell(string(comp.Impact)).
		SetTextColor(impactColor(comp.Impact)).SetExpansion(1))
	row++
	v.detail.SetCell(row, 0, detailLabelCell("Enabled"))
	v.detail.SetCell(row, 1, detailValueCell(enabled))
	row++
	v.detail.SetCell(row, 0, detailLabelCell("Uptime"))
	v.detail.SetCell(row, 1, detailValueCell(uptime))
	row++
	if comp.Description != "" {
		v.detail.SetCell(row, 0, detailLabelCell("Description"))
		v.detail.SetCell(row, 1, detailValueCell(comp.Description))
	}

	v.app.pages.SwitchToPage("compDetail")
	v.app.tv.SetFocus(v.detail)
}
