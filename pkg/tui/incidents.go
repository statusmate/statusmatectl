package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

// IncidentsView displays a list of incidents and handles creating/updating them.
type IncidentsView struct {
	app                *App
	table              *tview.Table
	detail             *tview.Table
	deleteModal        *tview.Modal
	detailCompIDs      []int
	detailCompRowStart int
	incidents          []api.Incident
	displayed          []api.Incident
	filterText         string
}

func newIncidentsView(app *App) *IncidentsView {
	v := &IncidentsView{app: app}

	v.table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSelectedStyle(tcell.StyleDefault.
			Background(tcell.ColorNavy).
			Foreground(tcell.ColorWhite))
	v.table.SetBorder(true)
	v.table.SetBackgroundColor(tcell.ColorBlack)
	v.table.SetTitle(" Incidents ")
	v.table.SetTitleAlign(tview.AlignCenter)
	v.table.SetInputCapture(v.onKey)

	v.detail = tview.NewTable().SetSelectable(true, false)
	v.detail.SetBorder(true)
	v.detail.SetTitle(" Incident Detail ")
	v.detail.SetTitleAlign(tview.AlignCenter)
	v.detail.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEscape:
			app.popPage()
			app.tv.SetFocus(v.table)
			return nil
		case tcell.KeyEnter:
			row, _ := v.detail.GetSelection()
			if v.detailCompRowStart > 0 && row >= v.detailCompRowStart {
				idx := row - v.detailCompRowStart
				if idx < len(v.detailCompIDs) {
					app.components.navigateTo(v.detailCompIDs[idx])
					app.switchTo(viewComponents)
				}
			}
			return nil
		}
		return ev
	})
	app.pages.AddPage("incDetail", v.detail, true, false)

	v.deleteModal = tview.NewModal().
		SetText("Delete this incident?").
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(idx int, label string) {
			app.popPage()
			app.tv.SetFocus(v.table)
			if label == "Delete" {
				v.confirmDelete()
			}
		})

	v.deleteModal.SetBackgroundColor(tcell.ColorBlack)
	v.deleteModal.SetBorder(true)
	app.pages.AddPage("incDelete", v.deleteModal, true, false)

	return v
}

func (v *IncidentsView) root() tview.Primitive { return v.table }

func (v *IncidentsView) refresh() {
	go func() {
		filter := api.PaginatedRequestFilter{}
		if v.app.statusPage != nil {
			filter["status_page"] = v.app.statusPage.ID
		}
		result, err := v.app.client.GetPaginatedIncidents(api.NewAllPaginatedRequest(filter))
		if err != nil {
			return
		}
		v.incidents = result.Results
		v.app.tv.QueueUpdateDraw(v.render)
	}()
}

func (v *IncidentsView) filter(text string) {
	v.filterText = text
	v.render()
}

func (v *IncidentsView) clearFilter() {
	v.filterText = ""
	v.render()
}

func (v *IncidentsView) render() {
	lower := strings.ToLower(v.filterText)
	v.displayed = v.displayed[:0]
	for _, inc := range v.incidents {
		if lower == "" ||
			strings.Contains(strings.ToLower(inc.Title), lower) ||
			strings.Contains(strings.ToLower(string(inc.Status)), lower) {
			v.displayed = append(v.displayed, inc)
		}
	}

	if lower != "" {
		v.table.SetTitle(fmt.Sprintf(" Incidents [%d/%d] [green::]</%s>[-:-:-] ", len(v.displayed), len(v.incidents), lower))
	} else {
		v.table.SetTitle(fmt.Sprintf(" Incidents [%d] ", len(v.incidents)))
	}
	v.table.Clear()

	for i, h := range []string{"UUID", "TITLE", "STATUS", "CREATED"} {
		v.table.SetCell(0, i, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false).
			SetExpansion(1))
	}

	for i, inc := range v.displayed {
		row := i + 1
		uuid := "-"
		if inc.UUID != nil {
			uuid = shortUUID(*inc.UUID)
		}
		v.table.SetCell(row, 0, tview.NewTableCell(uuid).SetTextColor(tcell.ColorGray))
		v.table.SetCell(row, 1, tview.NewTableCell(inc.Title).SetTextColor(tcell.ColorWhite).SetExpansion(3))
		v.table.SetCell(row, 2,
			tview.NewTableCell(formatIncidentStatus(inc.Status)).
				SetTextColor(incidentStatusColor(inc.Status)).
				SetExpansion(2))
		v.table.SetCell(row, 3, tview.NewTableCell(formatAge(inc.CreatedAt)).SetTextColor(tcell.ColorGray))
	}

	if len(v.displayed) > 0 {
		v.table.Select(1, 0)
	}
}

func (v *IncidentsView) selected() *api.Incident {
	row, _ := v.table.GetSelection()
	if row <= 0 || row-1 >= len(v.displayed) {
		return nil
	}
	return &v.displayed[row-1]
}

func (v *IncidentsView) onKey(ev *tcell.EventKey) *tcell.EventKey {
	switch ev.Key() {
	case tcell.KeyEnter:
		if inc := v.selected(); inc != nil {
			v.showDetail(inc)
		}
		return nil
	}
	switch ev.Rune() {
	case 'n':
		v.showCreateForm()
		return nil
	case 'u':
		if inc := v.selected(); inc != nil {
			v.showUpdateForm(inc)
		}
		return nil
	case 'd':
		if inc := v.selected(); inc != nil {
			v.showDeleteConfirm(inc)
		}
		return nil
	}
	return ev
}

func (v *IncidentsView) showDeleteConfirm(inc *api.Incident) {
	title := inc.Title
	v.deleteModal.SetText(fmt.Sprintf("Delete incident:\n[white::b]%s[-:-:-]?", title))
	v.app.pushPage("incDelete")
	v.app.tv.SetFocus(v.deleteModal)
}

func (v *IncidentsView) confirmDelete() {
	inc := v.selected()
	if inc == nil || inc.UUID == nil {
		return
	}
	uuid := *inc.UUID
	go func() {
		if err := v.app.client.DeleteIncident(uuid); err != nil {
			return
		}
		v.app.tv.QueueUpdateDraw(v.refresh)
	}()
}

func (v *IncidentsView) showDetail(inc *api.Incident) {
	v.detail.Clear()
	v.detailCompIDs = v.detailCompIDs[:0]
	v.detailCompRowStart = 0

	uuid := "-"
	if inc.UUID != nil {
		uuid = *inc.UUID
	}

	row := 0
	v.detail.SetCell(row, 0, detailSectionCell(inc.Title))
	row++
	v.detail.SetCell(row, 0, detailLabelCell("Status"))
	v.detail.SetCell(row, 1, tview.NewTableCell(formatIncidentStatus(inc.Status)).
		SetTextColor(incidentStatusColor(inc.Status)).SetExpansion(1))
	row++
	v.detail.SetCell(row, 0, detailLabelCell("UUID"))
	v.detail.SetCell(row, 1, detailValueCell(uuid))
	row++
	if inc.CreatedAt != nil {
		v.detail.SetCell(row, 0, detailLabelCell("Created"))
		v.detail.SetCell(row, 1, detailValueCell(inc.CreatedAt.Format("2006-01-02 15:04")))
		row++
	}
	if inc.Description != "" {
		v.detail.SetCell(row, 0, detailLabelCell("Description"))
		v.detail.SetCell(row, 1, detailValueCell(inc.Description))
		row++
	}
	if len(inc.Updates) > 0 {
		row++
		v.detail.SetCell(row, 0, detailSectionCell("Updates"))
		row++
		for _, u := range inc.Updates {
			v.detail.SetCell(row, 0, tview.NewTableCell(u.At.Format("2006-01-02 15:04")).
				SetTextColor(tcell.ColorGray))
			v.detail.SetCell(row, 1, tview.NewTableCell(formatIncidentStatus(u.Status)).
				SetTextColor(incidentStatusColor(u.Status)))
			v.detail.SetCell(row, 2, detailValueCell(u.Description))
			row++
		}
	}

	if inc.ID != nil {
		incID := *inc.ID
		cachedComps := make([]api.Component, len(v.app.components.components))
		copy(cachedComps, v.app.components.components)
		var statusPageFilter api.PaginatedRequestFilter
		if v.app.statusPage != nil {
			statusPageFilter = api.PaginatedRequestFilter{"status_page": v.app.statusPage.ID}
		}
		go func() {
			updatesResult, err := v.app.client.GetPaginatedUpdates(
				api.NewAllPaginatedRequest(api.PaginatedRequestFilter{"incident": incID}),
			)
			if err != nil {
				return
			}

			type compEntry struct {
				id     int
				impact api.ImpactType
				updAt  time.Time
			}
			byID := map[int]compEntry{}
			for _, u := range updatesResult.Results {
				for _, ac := range u.Components {
					existing, ok := byID[ac.Component]
					if !ok || u.At.After(existing.updAt) {
						byID[ac.Component] = compEntry{id: ac.Component, impact: ac.Impact, updAt: u.At}
					}
				}
			}
			if len(byID) == 0 {
				return
			}

			entries := make([]compEntry, 0, len(byID))
			ids := make([]int, 0, len(byID))
			for _, ce := range byID {
				entries = append(entries, ce)
				ids = append(ids, ce.id)
			}

			comps := cachedComps
			if len(comps) == 0 {
				filter := api.PaginatedRequestFilter{}
				if statusPageFilter != nil {
					filter = statusPageFilter
				}
				result, e := v.app.client.GetPaginatedComponents(api.NewAllPaginatedRequest(filter))
				if e == nil {
					comps = result.Results
				}
			}
			nameMap := make(map[int]string)
			for _, c := range comps {
				if c.ID != nil {
					nameMap[*c.ID] = c.Name
				}
			}

			v.app.tv.QueueUpdateDraw(func() {
				baseRow := v.detail.GetRowCount()
				baseRow++
				v.detail.SetCell(baseRow, 0, detailSectionCell("Affected Components"))
				baseRow++
				v.detailCompRowStart = baseRow
				v.detailCompIDs = ids
				for i, ce := range entries {
					name := fmt.Sprintf("id:%d", ce.id)
					if n, ok := nameMap[ce.id]; ok {
						name = n
					}
					v.detail.SetCell(baseRow+i, 0, tview.NewTableCell(name).
						SetTextColor(tcell.ColorWhite).SetExpansion(2))
					v.detail.SetCell(baseRow+i, 1, tview.NewTableCell(string(ce.impact)).
						SetTextColor(impactColor(ce.impact)).SetExpansion(1))
				}
			})
		}()
	}

	v.app.pushPage("incDetail")
	v.app.tv.SetFocus(v.detail)
}
