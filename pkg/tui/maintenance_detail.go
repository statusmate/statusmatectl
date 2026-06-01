package tui

import (
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

type MaintenanceDetailView struct {
	app   *App
	flex  *tview.Flex
	title *tview.TextView
	table *tview.Table
}

func newMaintenanceDetailView(app *App) *MaintenanceDetailView {
	d := &MaintenanceDetailView{app: app}

	d.title = tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true)
	d.title.SetBackgroundColor(tcell.ColorBlack)
	d.title.SetBorderPadding(0, 0, 0, 0)

	d.table = tview.NewTable().SetSelectable(true, false)
	d.table.SetBackgroundColor(tcell.ColorBlack)

	escHandler := func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			app.popPage()
		}
		return ev
	}
	d.title.SetInputCapture(escHandler)
	d.table.SetInputCapture(escHandler)

	d.flex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(d.title, 3, 0, false).
		AddItem(d.table, 0, 1, true)
	d.flex.SetBorder(true)
	d.flex.SetTitle(" Maintenance Detail ")
	d.flex.SetTitleAlign(tview.AlignCenter)
	d.flex.SetBackgroundColor(tcell.ColorBlack)

	app.pages.AddPage("maintDetail", d.flex, true, false)

	return d
}

func (d *MaintenanceDetailView) show(m *api.Maintenance) {
	d.table.Clear()
	d.title.SetText(m.Title)

	uuid := "-"
	if m.UUID != nil {
		uuid = *m.UUID
	}

	row := 0
	d.table.SetCell(row, 0, detailLabelCell("Status"))
	d.table.SetCell(row, 1, tview.NewTableCell(formatMaintenanceStatus(m.Status)).
		SetTextColor(maintenanceStatusColor(m.Status)).SetExpansion(1))
	row++
	d.table.SetCell(row, 0, detailLabelCell("UUID"))
	d.table.SetCell(row, 1, detailValueCell(uuid))
	row++
	d.table.SetCell(row, 0, detailLabelCell("Start"))
	d.table.SetCell(row, 1, detailValueCell(formatTimePtr(m.StartAt)))
	row++
	d.table.SetCell(row, 0, detailLabelCell("End"))
	d.table.SetCell(row, 1, detailValueCell(formatTimePtr(m.EndAt)))
	row++
	if m.Description != "" {
		d.table.SetCell(row, 0, detailLabelCell("Description"))
		d.table.SetCell(row, 1, detailValueCell(m.Description))
		row++
	}
	if len(m.Updates) > 0 {
		row++
		d.table.SetCell(row, 0, detailSectionCell("Updates"))
		row++
		for _, u := range m.Updates {
			d.table.SetCell(row, 0, tview.NewTableCell(u.At.Format("2006-01-02 15:04")).
				SetTextColor(tcell.ColorGray))
			d.table.SetCell(row, 1, tview.NewTableCell(formatMaintenanceStatus(u.Status)).
				SetTextColor(maintenanceStatusColor(u.Status)))
			d.table.SetCell(row, 2, detailValueCell(u.Description))
			row++
		}
	}

	d.app.pushPage("maintDetail")
	d.app.tv.SetFocus(d.table)
}
