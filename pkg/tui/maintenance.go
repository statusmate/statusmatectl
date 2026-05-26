package tui

import (
	"fmt"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

// MaintenanceView displays a list of maintenance windows.
type MaintenanceView struct {
	app          *App
	table        *tview.Table
	detail       *tview.Table
	maintenances []api.Maintenance
}

func newMaintenanceView(app *App) *MaintenanceView {
	v := &MaintenanceView{app: app}

	v.table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSelectedStyle(tcell.StyleDefault.
			Background(tcell.ColorNavy).
			Foreground(tcell.ColorWhite))
	v.table.SetBorder(true)
	v.table.SetTitle(" Maintenance ")
	v.table.SetTitleAlign(tview.AlignCenter)
	v.table.SetInputCapture(v.onKey)

	v.detail = tview.NewTable().SetSelectable(true, false)
	v.detail.SetBorder(true)
	v.detail.SetTitle(" Maintenance Detail ")
	v.detail.SetTitleAlign(tview.AlignCenter)
	v.detail.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			app.pages.SwitchToPage(viewMaintenance)
			app.tv.SetFocus(v.table)
		}
		return ev
	})
	app.pages.AddPage("maintDetail", v.detail, true, false)

	return v
}

func (v *MaintenanceView) root() tview.Primitive { return v.table }

func (v *MaintenanceView) refresh() {
	go func() {
		filter := api.PaginatedRequestFilter{}
		if v.app.statusPage != nil {
			filter["status_page"] = v.app.statusPage.ID
		}
		result, err := v.app.client.GetPaginatedMaintenance(api.NewAllPaginatedRequest(filter))
		if err != nil {
			return
		}
		v.maintenances = result.Results
		v.app.tv.QueueUpdateDraw(v.render)
	}()
}

func (v *MaintenanceView) render() {
	v.table.SetTitle(fmt.Sprintf(" Maintenance [%d] ", len(v.maintenances)))
	v.table.Clear()

	for i, h := range []string{"UUID", "TITLE", "STATUS", "START", "END"} {
		v.table.SetCell(0, i, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false).
			SetExpansion(1))
	}

	for i, m := range v.maintenances {
		row := i + 1
		uuid := "-"
		if m.UUID != nil {
			uuid = shortUUID(*m.UUID)
		}
		v.table.SetCell(row, 0, tview.NewTableCell(uuid).SetTextColor(tcell.ColorGray))
		v.table.SetCell(row, 1, tview.NewTableCell(m.Title).SetTextColor(tcell.ColorWhite).SetExpansion(3))
		v.table.SetCell(row, 2,
			tview.NewTableCell(formatMaintenanceStatus(m.Status)).
				SetTextColor(maintenanceStatusColor(m.Status)).
				SetExpansion(2))
		v.table.SetCell(row, 3, tview.NewTableCell(formatTimePtr(m.StartAt)).SetTextColor(tcell.ColorGray))
		v.table.SetCell(row, 4, tview.NewTableCell(formatTimePtr(m.EndAt)).SetTextColor(tcell.ColorGray))
	}

	if len(v.maintenances) > 0 {
		v.table.Select(1, 0)
	}
}

func (v *MaintenanceView) selected() *api.Maintenance {
	row, _ := v.table.GetSelection()
	if row <= 0 || row-1 >= len(v.maintenances) {
		return nil
	}
	return &v.maintenances[row-1]
}

func (v *MaintenanceView) onKey(ev *tcell.EventKey) *tcell.EventKey {
	if ev.Key() == tcell.KeyEnter {
		if m := v.selected(); m != nil {
			v.showDetail(m)
		}
		return nil
	}
	return ev
}

func (v *MaintenanceView) showDetail(m *api.Maintenance) {
	v.detail.Clear()

	uuid := "-"
	if m.UUID != nil {
		uuid = *m.UUID
	}

	row := 0
	v.detail.SetCell(row, 0, detailSectionCell(m.Title))
	row++
	v.detail.SetCell(row, 0, detailLabelCell("Status"))
	v.detail.SetCell(row, 1, tview.NewTableCell(formatMaintenanceStatus(m.Status)).
		SetTextColor(maintenanceStatusColor(m.Status)).SetExpansion(1))
	row++
	v.detail.SetCell(row, 0, detailLabelCell("UUID"))
	v.detail.SetCell(row, 1, detailValueCell(uuid))
	row++
	v.detail.SetCell(row, 0, detailLabelCell("Start"))
	v.detail.SetCell(row, 1, detailValueCell(formatTimePtr(m.StartAt)))
	row++
	v.detail.SetCell(row, 0, detailLabelCell("End"))
	v.detail.SetCell(row, 1, detailValueCell(formatTimePtr(m.EndAt)))
	row++
	if m.Description != "" {
		v.detail.SetCell(row, 0, detailLabelCell("Description"))
		v.detail.SetCell(row, 1, detailValueCell(m.Description))
		row++
	}
	if len(m.Updates) > 0 {
		row++
		v.detail.SetCell(row, 0, detailSectionCell("Updates"))
		row++
		for _, u := range m.Updates {
			v.detail.SetCell(row, 0, tview.NewTableCell(u.At.Format("2006-01-02 15:04")).
				SetTextColor(tcell.ColorGray))
			v.detail.SetCell(row, 1, tview.NewTableCell(formatMaintenanceStatus(u.Status)).
				SetTextColor(maintenanceStatusColor(u.Status)))
			v.detail.SetCell(row, 2, detailValueCell(u.Description))
			row++
		}
	}

	v.app.pages.SwitchToPage("maintDetail")
	v.app.tv.SetFocus(v.detail)
}
