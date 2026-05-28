package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
	"github.com/statusmate/statusmatectl/pkg/editor"
	"github.com/statusmate/statusmatectl/pkg/format"
)

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
					v.app.pages.RemovePage("confirm-create-incident")
					confirmed <- (label == "Create")
				})
			v.app.pages.AddPage("confirm-create-incident", modal, true, true)
			v.app.tv.SetFocus(modal)
		})

		if <-confirmed {
			v.app.client.CreateIncident(payload) //nolint:errcheck
		}

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
		data += api.BuildComponentsEditorFooter(availableComponents)

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
