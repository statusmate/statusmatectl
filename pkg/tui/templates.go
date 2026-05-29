package tui

import (
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/editor"
	"github.com/statusmate/statusmatectl/pkg/format"
)

// TemplatesView displays incident/maintenance templates for a status page.
type TemplatesView struct {
	app        *App
	table      *tview.Table
	detail     *tview.TextView
	templates  []api.Template
	displayed  []api.Template
	filterText string
}

func newTemplatesView(app *App) *TemplatesView {
	v := &TemplatesView{app: app}

	v.table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSelectedStyle(tcell.StyleDefault.
			Background(tcell.ColorNavy).
			Foreground(tcell.ColorWhite))
	v.table.SetBorder(true)
	v.table.SetTitle(" Templates ")
	v.table.SetTitleAlign(tview.AlignCenter)
	v.table.SetInputCapture(v.onKey)

	v.detail = tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	v.detail.SetBorder(true)
	v.detail.SetTitle(" Template Detail ")
	v.detail.SetTitleAlign(tview.AlignCenter)
	v.detail.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			app.popPage()
			app.tv.SetFocus(v.table)
		}
		return ev
	})
	app.pages.AddPage("tmplDetail", v.detail, true, false)

	return v
}

func (v *TemplatesView) root() tview.Primitive { return v.table }

func (v *TemplatesView) refresh() {
	go func() {
		filter := api.PaginatedRequestFilter{}
		if v.app.statusPage != nil {
			filter["status_page"] = v.app.statusPage.ID
		}
		result, err := v.app.client.GetPaginatedTemplates(api.NewAllPaginatedRequest(filter))
		if err != nil {
			return
		}
		v.templates = result.Results
		v.app.tv.QueueUpdateDraw(v.render)
	}()
}

func (v *TemplatesView) filter(text string) {
	v.filterText = text
	v.render()
}

func (v *TemplatesView) clearFilter() {
	v.filterText = ""
	v.render()
}

func (v *TemplatesView) render() {
	lower := strings.ToLower(v.filterText)
	v.displayed = v.displayed[:0]
	for _, t := range v.templates {
		if lower == "" ||
			strings.Contains(strings.ToLower(t.Title), lower) ||
			strings.Contains(strings.ToLower(t.FriendlyName), lower) ||
			strings.Contains(strings.ToLower(t.Description), lower) {
			v.displayed = append(v.displayed, t)
		}
	}

	if lower != "" {
		v.table.SetTitle(fmt.Sprintf(" Templates [%d/%d] [green::]</%s>[-:-:-] ", len(v.displayed), len(v.templates), lower))
	} else {
		v.table.SetTitle(fmt.Sprintf(" Templates [%d] ", len(v.templates)))
	}
	v.table.Clear()

	for i, h := range []string{"UUID", "TITLE", "STATUS", "COMPONENTS"} {
		v.table.SetCell(0, i, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false).
			SetExpansion(1))
	}

	for i, t := range v.displayed {
		row := i + 1
		uuid := "-"
		if t.UUID != nil {
			uuid = shortUUID(*t.UUID)
		}
		status := "-"
		statusColor := tcell.ColorGray
		if t.Status != nil && *t.Status != "" {
			status = prettyStatus(*t.Status)
			statusColor = templateStatusColor(*t.Status)
		}
		compCount := fmt.Sprintf("%d", len(t.Components))

		v.table.SetCell(row, 0, tview.NewTableCell(uuid).SetTextColor(tcell.ColorGray))
		v.table.SetCell(row, 1, tview.NewTableCell(t.Title).SetTextColor(tcell.ColorWhite).SetExpansion(3))
		v.table.SetCell(row, 2, tview.NewTableCell(status).SetTextColor(statusColor).SetExpansion(2))
		v.table.SetCell(row, 3, tview.NewTableCell(compCount).SetTextColor(tcell.ColorGray))
	}

	if len(v.displayed) > 0 {
		v.table.Select(1, 0)
	}
}

func (v *TemplatesView) selected() *api.Template {
	row, _ := v.table.GetSelection()
	if row <= 0 || row-1 >= len(v.displayed) {
		return nil
	}
	return &v.displayed[row-1]
}

func (v *TemplatesView) onKey(ev *tcell.EventKey) *tcell.EventKey {
	if ev.Key() == tcell.KeyEnter {
		if t := v.selected(); t != nil {
			v.showDetail(t)
		}
		return nil
	}
	switch ev.Rune() {
	case 'd':
		if t := v.selected(); t != nil {
			v.showDeleteConfirm(t)
		}
		return nil
	case 'n':
		if t := v.selected(); t != nil {
			v.showCreateFromTemplate(t)
		}
		return nil
	}
	return ev
}

func (v *TemplatesView) showDeleteConfirm(t *api.Template) {
	if t.UUID == nil {
		return
	}
	uuid := *t.UUID
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Delete template %q?", t.Title)).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(_ int, label string) {
			v.app.pages.RemovePage("confirm-delete-template")
			if label != "Delete" {
				return
			}
			go func() {
				v.app.client.DeleteTemplate(uuid) //nolint:errcheck
				v.refresh()
			}()
		})
	v.app.pages.AddPage("confirm-delete-template", modal, true, true)
	v.app.tv.SetFocus(modal)
}

func (v *TemplatesView) showCreateFromTemplate(t *api.Template) {
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
		payload.Title = t.Title
		payload.Description = t.Description
		payload.Notify = t.Notify
		if t.Status != nil && *t.Status != "" {
			payload.Status = *t.Status
		}
		payload.Components = affectedComponentsToStrings(t.Components, comps.Results)

		data, err := format.Marshal(payload, &api.CreateIncidentPayloadFieldDescriptions)
		if err != nil {
			return
		}
		data += api.BuildComponentsEditorFooter(comps.Results)

		v.app.tv.Suspend(func() {
			output, err := editor.CaptureInputFromEditor([]byte(data))
			if err != nil {
				return
			}
			if err := format.Unmarshal(string(output), payload); err != nil {
				return
			}
		})

		if strings.TrimSpace(payload.Title) == "" {
			return
		}

		confirmed := make(chan bool, 1)
		v.app.tv.QueueUpdateDraw(func() {
			modal := tview.NewModal().
				SetText(fmt.Sprintf("Create incident: %q?", payload.Title)).
				AddButtons([]string{"Create", "Cancel"}).
				SetDoneFunc(func(_ int, label string) {
					v.app.pages.RemovePage("confirm-create-template-incident")
					confirmed <- (label == "Create")
				})
			v.app.pages.AddPage("confirm-create-template-incident", modal, true, true)
			v.app.tv.SetFocus(modal)
		})

		if <-confirmed {
			v.app.client.CreateIncident(payload) //nolint:errcheck
		}
	}()
}

func (v *TemplatesView) showDetail(t *api.Template) {
	v.detail.Clear()

	uuid := "-"
	if t.UUID != nil {
		uuid = *t.UUID
	}

	fmt.Fprintf(v.detail, "[yellow::b]%s[-:-:-]\n\n", t.Title)
	if t.FriendlyName != "" {
		fmt.Fprintf(v.detail, "[blue]Friendly name:[-]  %s\n", t.FriendlyName)
	}
	fmt.Fprintf(v.detail, "[blue]UUID:[-]  %s\n", uuid)
	if t.Status != nil && *t.Status != "" {
		fmt.Fprintf(v.detail, "[blue]Status:[-]  [%s]%s[-]\n", colorTag(templateStatusColor(*t.Status)), prettyStatus(*t.Status))
	}
	fmt.Fprintf(v.detail, "[blue]Notify:[-]  %v\n", t.Notify)
	if t.CreatedAt != nil {
		fmt.Fprintf(v.detail, "[blue]Created:[-]  %s\n", t.CreatedAt.Format("2006-01-02 15:04"))
	}
	if t.Description != "" {
		fmt.Fprintf(v.detail, "\n[blue]Description:[-]\n%s\n", t.Description)
	}
	if len(t.Components) > 0 {
		fmt.Fprintf(v.detail, "\n[blue]Components:[-]\n")
		for _, c := range t.Components {
			fmt.Fprintf(v.detail, "  [%s]%s[-]  id:%d\n",
				colorTag(impactColor(c.Impact)),
				string(c.Impact),
				c.Component,
			)
		}
	}
	if len(t.AssignedTags) > 0 {
		fmt.Fprintf(v.detail, "\n[blue]Tags:[-]\n")
		for _, tag := range t.AssignedTags {
			fmt.Fprintf(v.detail, "  %s\n", tag.Title)
		}
	}

	v.app.pushPage("tmplDetail")
	v.app.tv.SetFocus(v.detail)
}

func templateStatusColor(status string) tcell.Color {
	switch api.IncidentStatusType(status) {
	case api.IncidentStatusResolved:
		return tcell.ColorGreen
	case api.IncidentStatusInvestigation:
		return tcell.ColorRed
	case api.IncidentStatusMonitoring:
		return tcell.ColorYellow
	case api.IncidentStatusIdentified:
		return tcell.ColorOrange
	}
	switch api.MaintenanceStatusType(status) {
	case api.MaintenanceStatusCompleted:
		return tcell.ColorGreen
	case api.MaintenanceStatusInProgress:
		return tcell.ColorYellow
	case api.MaintenanceStatusNotStarted:
		return tcell.ColorGray
	}
	return tcell.ColorWhite
}
