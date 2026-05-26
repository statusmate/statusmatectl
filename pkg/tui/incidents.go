package tui

import (
	"fmt"
	"time"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

// IncidentsView displays a list of incidents and handles creating/updating them.
type IncidentsView struct {
	app       *App
	table     *tview.Table
	incidents []api.Incident
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

	v.table.SetInputCapture(v.onKey)
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

func (v *IncidentsView) render() {
	v.table.Clear()

	for i, h := range []string{"UUID", "TITLE", "STATUS", "CREATED"} {
		v.table.SetCell(0, i, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false).
			SetExpansion(1))
	}

	for i, inc := range v.incidents {
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

	if len(v.incidents) > 0 {
		v.table.Select(1, 0)
	}
}

func (v *IncidentsView) selected() *api.Incident {
	row, _ := v.table.GetSelection()
	if row <= 0 || row-1 >= len(v.incidents) {
		return nil
	}
	return &v.incidents[row-1]
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
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)

	uuid := "-"
	if inc.UUID != nil {
		uuid = *inc.UUID
	}

	text := fmt.Sprintf("[yellow::b]%s[-:-:-]\n\n", inc.Title)
	text += fmt.Sprintf("[blue]Status:[-]  %s\n", formatIncidentStatus(inc.Status))
	text += fmt.Sprintf("[blue]UUID:[-]    %s\n", uuid)
	if inc.CreatedAt != nil {
		text += fmt.Sprintf("[blue]Created:[-] %s\n", inc.CreatedAt.Format("2006-01-02 15:04"))
	}
	if inc.Description != "" {
		text += fmt.Sprintf("\n[blue]Description:[-]\n%s\n", inc.Description)
	}
	if len(inc.Updates) > 0 {
		text += "\n[yellow]Updates:[-]\n"
		for _, u := range inc.Updates {
			text += fmt.Sprintf("  [gray]%s[-]  [aqua]%s[-]\n  %s\n\n",
				u.At.Format("2006-01-02 15:04"),
				formatIncidentStatus(u.Status),
				u.Description)
		}
	}

	tv.SetText(text)
	tv.SetBorder(true).SetTitle(" Incident Detail ").SetTitleAlign(tview.AlignLeft)
	tv.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			v.app.closeModal("incDetail")
		}
		return ev
	})
	v.app.showModal("incDetail", tv, 80, 30)
}

func (v *IncidentsView) showCreateForm() {
	if v.app.statusPage == nil {
		return
	}

	form := tview.NewForm()
	var title, description string
	statusOpts := incidentStatusOptions()
	statusIdx := 0

	form.AddInputField("Title*", "", 50, nil, func(t string) { title = t })
	form.AddInputField("Description", "", 50, nil, func(t string) { description = t })
	form.AddDropDown("Status", statusOpts, 0, func(_ string, i int) { statusIdx = i })
	form.AddButton("Create", func() {
		if title == "" {
			return
		}
		payload := api.NewCreateIncidentPayload(v.app.statusPage)
		payload.Title = title
		payload.Description = description
		payload.Status = statusOpts[statusIdx]
		v.app.closeModal("incCreate")
		go func() {
			v.app.client.CreateIncident(payload)
			v.refresh()
		}()
	})
	form.AddButton("Cancel", func() {
		v.app.closeModal("incCreate")
	})
	form.SetBorder(true).SetTitle(" New Incident ").SetTitleAlign(tview.AlignLeft)
	form.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			v.app.closeModal("incCreate")
			return nil
		}
		return ev
	})
	v.app.showModal("incCreate", form, 60, 15)
}

func (v *IncidentsView) showUpdateForm(inc *api.Incident) {
	form := tview.NewForm()
	statusOpts := incidentStatusOptions()
	statusIdx := 0
	for i, s := range statusOpts {
		if s == string(inc.Status) {
			statusIdx = i
			break
		}
	}
	var message string

	form.AddDropDown("Status", statusOpts, statusIdx, func(_ string, i int) { statusIdx = i })
	form.AddInputField("Message", "", 50, nil, func(t string) { message = t })
	form.AddButton("Add Update", func() {
		if inc.ID == nil {
			return
		}
		update := &api.IncidentUpdate{
			Incident:    inc.ID,
			Status:      api.IncidentStatusType(statusOpts[statusIdx]),
			Description: message,
			Notify:      true,
			At:          time.Now(),
			Components:  make([]api.AffectedComponent, 0),
		}
		v.app.closeModal("incUpdate")
		go func() {
			v.app.client.CreateIncidentUpdate(update)
			v.refresh()
		}()
	})
	form.AddButton("Cancel", func() {
		v.app.closeModal("incUpdate")
	})
	form.SetBorder(true).
		SetTitle(fmt.Sprintf(" Update: %s ", truncate(inc.Title, 30))).
		SetTitleAlign(tview.AlignLeft)
	form.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			v.app.closeModal("incUpdate")
			return nil
		}
		return ev
	})
	v.app.showModal("incUpdate", form, 60, 13)
}

func incidentStatusOptions() []string {
	var opts []string
	for _, s := range api.IncidentStatusList() {
		opts = append(opts, string(s))
	}
	return opts
}
