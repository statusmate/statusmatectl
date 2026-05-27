package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/editor"
	"github.com/statusmate/statusmatectl/pkg/format"
)

// IncidentsView displays a list of incidents and handles creating/updating them.
type IncidentsView struct {
	app                *App
	table              *tview.Table
	detail             *tview.Table
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
			app.pages.SwitchToPage(viewIncidents)
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
	}
	return ev
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

	v.app.pages.SwitchToPage("incDetail")
	v.app.tv.SetFocus(v.detail)
}

func (v *IncidentsView) showCreateForm() {
	if v.app.statusPage == nil {
		return
	}

	go func() {
		comps, err := v.app.client.GetPaginatedComponents(
			api.NewAllPaginatedRequest(api.PaginatedRequestFilter{"status_page": v.app.statusPage.ID}),
		)
		if err != nil {
			return
		}

		payload := api.NewCreateIncidentPayload(v.app.statusPage)

		data, err := format.Marshal(payload, &api.CreateIncidentPayloadFieldDescriptions)
		if err != nil {
			return
		}
		data += buildComponentsFooter(comps.Results)

		v.app.tv.Suspend(func() {
			output, err := editor.CaptureInputFromEditor([]byte(data))
			if err != nil {
				return
			}
			if err := format.Unmarshal(string(output), payload); err != nil {
				return
			}
			if strings.TrimSpace(payload.Title) == "" {
				return
			}
			v.app.client.CreateIncident(payload) //nolint:errcheck
		})

		v.refresh()
	}()
}

func (v *IncidentsView) showUpdateForm(inc *api.Incident) {
	if inc.ID == nil {
		return
	}

	go func() {
		filter := api.PaginatedRequestFilter{}
		if v.app.statusPage != nil {
			filter["status_page"] = v.app.statusPage.ID
		}
		comps, err := v.app.client.GetPaginatedComponents(api.NewAllPaginatedRequest(filter))
		if err != nil {
			return
		}
		availableComponents := comps.Results

		latestUpdate, _ := v.app.client.GetLatestIncidentUpdate(*inc.ID)

		var sourceComponents []api.AffectedComponent
		if latestUpdate != nil {
			sourceComponents = latestUpdate.Components
		} else {
			sourceComponents = inc.Components
		}

		payload := &api.CreateIncidentUpdatePayload{
			Status:     string(inc.Status),
			Components: affectedComponentsToStrings(sourceComponents, availableComponents),
			Notify:     true,
		}

		data, err := format.Marshal(payload, &api.CreateIncidentUpdatePayloadFieldDescriptions)
		if err != nil {
			return
		}
		data += buildComponentsFooter(availableComponents)

		v.app.tv.Suspend(func() {
			output, err := editor.CaptureInputFromEditor([]byte(data))
			if err != nil {
				return
			}
			if err := format.Unmarshal(string(output), payload); err != nil {
				return
			}
			if strings.TrimSpace(payload.Description) == "" {
				return
			}
			affectedComps, err := api.BuildAffectedComponents(payload.Components, availableComponents)
			if err != nil {
				return
			}
			update := &api.IncidentUpdate{
				Incident:    inc.ID,
				Status:      api.IncidentStatusType(payload.Status),
				Description: payload.Description,
				Notify:      payload.Notify,
				At:          time.Now(),
				Components:  affectedComps,
			}
			v.app.client.CreateIncidentUpdate(update) //nolint:errcheck
		})

		v.refresh()
	}()
}

func buildComponentsFooter(components []api.Component) string {
	if len(components) == 0 {
		return ""
	}

	children := make(map[int][]api.Component)
	var roots []api.Component
	for _, c := range components {
		if c.Parent == nil {
			roots = append(roots, c)
		} else {
			children[*c.Parent] = append(children[*c.Parent], c)
		}
	}

	var sb strings.Builder
	sb.WriteString("\n# Доступные компоненты:\n")

	var write func(comps []api.Component, indent string)
	write = func(comps []api.Component, indent string) {
		for _, c := range comps {
			sb.WriteString(fmt.Sprintf("#%s- %s\n", indent, c.Name))
			if c.ID != nil {
				if kids, ok := children[*c.ID]; ok {
					write(kids, indent+"  ")
				}
			}
		}
	}
	write(roots, " ")

	return sb.String()
}

func affectedComponentsToStrings(comps []api.AffectedComponent, available []api.Component) []string {
	result := make([]string, 0, len(comps))
	for _, ac := range comps {
		var name string
		for _, c := range available {
			if c.ID != nil && *c.ID == ac.Component {
				name = c.Name
				break
			}
		}
		if name == "" {
			continue
		}
		result = append(result, fmt.Sprintf("%s %s", ac.Impact, name))
	}
	return result
}

