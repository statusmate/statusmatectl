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

	v.table.SetInputCapture(v.onKey)
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
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)

	uuid := "-"
	if m.UUID != nil {
		uuid = *m.UUID
	}

	text := fmt.Sprintf("[yellow::b]%s[-:-:-]\n\n", m.Title)
	text += fmt.Sprintf("[blue]Status:[-]  %s\n", formatMaintenanceStatus(m.Status))
	text += fmt.Sprintf("[blue]UUID:[-]    %s\n", uuid)
	text += fmt.Sprintf("[blue]Start:[-]   %s\n", formatTimePtr(m.StartAt))
	text += fmt.Sprintf("[blue]End:[-]     %s\n", formatTimePtr(m.EndAt))
	if m.Description != "" {
		text += fmt.Sprintf("\n[blue]Description:[-]\n%s\n", m.Description)
	}
	if len(m.Updates) > 0 {
		text += "\n[yellow]Updates:[-]\n"
		for _, u := range m.Updates {
			text += fmt.Sprintf("  [gray]%s[-]  [aqua]%s[-]\n  %s\n\n",
				u.At.Format("2006-01-02 15:04"),
				formatMaintenanceStatus(u.Status),
				u.Description)
		}
	}

	tv.SetText(text)
	tv.SetBorder(true).SetTitle(" Maintenance Detail ").SetTitleAlign(tview.AlignLeft)
	tv.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			v.app.closeModal("maintDetail")
		}
		return ev
	})
	v.app.showModal("maintDetail", tv, 80, 30)
}
