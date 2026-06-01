package tui

import (
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

// MaintenanceView displays a list of maintenance windows.
type MaintenanceView struct {
	app          *App
	table        *tview.Table
	detail       *MaintenanceDetailView
	describe     *MaintenanceDescribeView
	deleteModal  *tview.Modal
	maintenances []api.Maintenance
	displayed    []api.Maintenance
	filterText   string
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
	v.table.SetTitle(" Maintenances ")
	v.table.SetBackgroundColor(tcell.ColorBlack)
	v.table.SetTitleAlign(tview.AlignCenter)
	v.table.SetInputCapture(v.onKey)

	v.table.SetDrawFunc(func(_ tcell.Screen, _, _, width, height int) (int, int, int, int) {
		// uuid(6) + status(20) + start(20) + end(20) + separators(5) = 70
		const fixedCols = 70
		titleMax := max(width - fixedCols, 10)

		for row := 1; row < v.table.GetRowCount(); row++ {
			if cell := v.table.GetCell(row, 1); cell != nil {
				cell.SetMaxWidth(titleMax)
			}
		}

		for i, m := range v.displayed {
			if cell := v.table.GetCell(i+1, 1); cell != nil {
				cell.SetText(truncate(m.Title, titleMax))
			}
		}

		return v.table.GetInnerRect()
	})

	v.detail = newMaintenanceDetailView(app)
	v.describe = newMaintenanceDescribeView(app)

	v.deleteModal = tview.NewModal().
		SetText("Delete this maintenance?").
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(idx int, label string) {
			app.popPage()
			app.tv.SetFocus(v.table)
			if label == "Delete" {
				v.confirmDelete()
			}
		})
	app.pages.AddPage("maintDelete", v.deleteModal, true, false)

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

func (v *MaintenanceView) filter(text string) {
	v.filterText = text
	v.render()
}

func (v *MaintenanceView) clearFilter() {
	v.filterText = ""
	v.render()
}

func (v *MaintenanceView) render() {
	lower := strings.ToLower(v.filterText)
	v.displayed = v.displayed[:0]
	for _, m := range v.maintenances {
		if lower == "" ||
			strings.Contains(strings.ToLower(m.Title), lower) ||
			strings.Contains(strings.ToLower(string(m.Status)), lower) {
			v.displayed = append(v.displayed, m)
		}
	}

	if lower != "" {
		v.table.SetTitle(fmt.Sprintf(" Maintenances [%d/%d] ", len(v.displayed), len(v.maintenances)))
	} else {
		v.table.SetTitle(fmt.Sprintf(" Maintenances [%d] ", len(v.maintenances)))
	}
	v.table.Clear()

	for i, h := range []string{"UUID", "TITLE", "STATUS", "START", "END"} {
		v.table.SetCell(0, i, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false).
			SetExpansion(1))
	}

	for i, m := range v.displayed {
		row := i + 1
		uuid := "-"
		if m.UUID != nil {
			uuid = shortUUID(*m.UUID)
		}
		v.table.SetCell(row, 0, tview.NewTableCell(uuid).SetTextColor(tcell.ColorGray))
		v.table.SetCell(row, 1, tview.NewTableCell(m.Title).SetTextColor(tcell.ColorWhite))
		v.table.SetCell(row, 2,
			tview.NewTableCell(formatMaintenanceStatus(m.Status)).
				SetTextColor(maintenanceStatusColor(m.Status)))
		v.table.SetCell(row, 3, tview.NewTableCell(formatTimePtr(m.StartAt)).SetTextColor(tcell.ColorGray))
		v.table.SetCell(row, 4, tview.NewTableCell(formatTimePtr(m.EndAt)).SetTextColor(tcell.ColorGray))
	}

	if len(v.displayed) > 0 {
		v.table.Select(1, 0)
	}
}

func (v *MaintenanceView) selected() *api.Maintenance {
	row, _ := v.table.GetSelection()
	if row <= 0 || row-1 >= len(v.displayed) {
		return nil
	}
	return &v.displayed[row-1]
}

func (v *MaintenanceView) onKey(ev *tcell.EventKey) *tcell.EventKey {
	if ev.Key() == tcell.KeyEnter {
		if m := v.selected(); m != nil {
			v.describe.show(m)
		}
		return nil
	}
	if ev.Rune() == 'd' {
		if m := v.selected(); m != nil {
			v.showDeleteConfirm(m)
		}
		return nil
	}
	return ev
}

func (v *MaintenanceView) showDeleteConfirm(m *api.Maintenance) {
	v.deleteModal.SetText(fmt.Sprintf("Delete maintenance:\n[white::b]%s[-:-:-]?", m.Title))
	v.app.pushPage("maintDelete")
	v.app.tv.SetFocus(v.deleteModal)
}

func (v *MaintenanceView) confirmDelete() {
	m := v.selected()
	if m == nil || m.UUID == nil {
		return
	}
	uuid := *m.UUID
	go func() {
		if err := v.app.client.DeleteMaintenance(uuid); err != nil {
			return
		}
		v.app.tv.QueueUpdateDraw(v.refresh)
	}()
}

func (v *MaintenanceView) showDetail(m *api.Maintenance) {
	v.detail.show(m)
}
