package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type IncidentStatusType string

const (
	IncidentStatusInvestigation IncidentStatusType = "incident_investigating"
	IncidentStatusIdentified    IncidentStatusType = "incident_identified"
	IncidentStatusMonitoring    IncidentStatusType = "incident_monitoring"
	IncidentStatusResolved      IncidentStatusType = "incident_resolved"
	IncidentStatusNotice        IncidentStatusType = "notice"
)

var CommentsMap = map[string]string{
	"name": "Название инцидента",
	"desc": "Описание инцидента",
	"status": `Available statuses:
- incident_investigating
- incident_identified
- incident_monitoring
- incident_resolved`,
	"affected": `Impacts:
- o, op - operational
- u, um - under maintenance
- d, dp - degraded performance
- p, po - partial outage
- m, mo - major outage`,
}

func IncidentStatusList() []IncidentStatusType {
	return []IncidentStatusType{
		IncidentStatusInvestigation,
		IncidentStatusIdentified,
		IncidentStatusMonitoring,
		IncidentStatusResolved,
		IncidentStatusNotice,
	}
}

func IncidentActiveStatusList() []IncidentStatusType {
	return []IncidentStatusType{
		IncidentStatusInvestigation,
		IncidentStatusIdentified,
		IncidentStatusMonitoring,
	}
}

type Incident struct {
	ID           *int                `json:"id,omitempty"`
	UUID         *string             `json:"uuid,omitempty" tab:"UUID"`
	AbsoluteURL  *string             `json:"absolute_url,omitempty"`
	Title        string              `json:"title" tab:"Title"`
	Status       IncidentStatusType  `json:"status" tab:"Status"`
	Components   []AffectedComponent `json:"components"`
	Notify       bool                `json:"notify"`
	CreatedAt    *time.Time          `json:"created_at,omitempty"`
	CreatedBy    *int                `json:"created_by,omitempty"`
	UpdatedAt    *time.Time          `json:"updated_at,omitempty"`
	LastUpdateAt *time.Time          `json:"last_update_at,omitempty"`
	Description  string              `json:"description"`
	EndAt        *time.Time          `json:"end_at,omitempty"`
	Logs         []Log               `json:"logs,omitempty"`
	Updates      []IncidentUpdate    `json:"updates,omitempty"`
	StartAt      time.Time           `json:"start_at"`
	StatusPage   int                 `json:"status_page"`
	PrivateNote  string              `json:"private_note"`
	ShowOnTop    bool                `json:"show_on_top"`
	AffectUptime bool                `json:"affect_uptime"`
}

type CreateIncidentPayload struct {
	StartAt      time.Time
	Title        string    `format:"title"`
	Description  string    `format:"description"`
	Status       string    `format:"status"`
	Components   []string  `format:"components"`
	Notify       bool      `format:"notify"`
	ShowOnTop    bool      `format:"show_on_top"`
	AffectUptime bool      `format:"affect_uptime"`
	PrivateNote  string    `format:"private_note"`
	StatusPage   int
}

var CreateIncidentPayloadFieldDescriptions = map[string]string{
	"title":       "Название инцидента",
	"description": "Описание инцидента",
	"status": `Возможные статусы:
- incident_investigating
- incident_identified
- incident_monitoring
- incident_resolved`,
	"components": `Указываются в формате: [Impact] [Имя компонента]
Возможные значения Impact:
- o, op  — operational
- u, um  — under maintenance
- d, dp  — degraded performance
- p, po  — partial outage
- m, mo  — major outage

Примеры:
p Web
m API`,
	"private_note":  "Приватное примечание, не отображается публично",
	"start_at":      "Время инцидента",
	"notify":        "Отправлять ли уведомление пользователям",
	"show_on_top":   "Отображать ли инцидент выше других",
	"affect_uptime": "Влияет ли инцидент на аптайм компонента",
}

func NewIncident(statusPage *StatusPage) *Incident {
	return &Incident{
		StatusPage:   statusPage.ID,
		Status:       IncidentStatusInvestigation,
		Components:   []AffectedComponent{},
		Title:        "",
		Description:  "",
		PrivateNote:  "",
		Notify:       true,
		StartAt:      time.Now(),
		ShowOnTop:    true,
		AffectUptime: true,
	}
}

func NewCreateIncidentPayload(statusPage *StatusPage) *CreateIncidentPayload {
	return &CreateIncidentPayload{
		StatusPage:   statusPage.ID,
		Status:       string(IncidentStatusInvestigation),
		Components:   []string{},
		Title:        "",
		Description:  "",
		PrivateNote:  "",
		Notify:       true,
		ShowOnTop:    true,
		AffectUptime: true,
		StartAt:      time.Now(),
	}
}

func (c *Client) CreateIncident(input *CreateIncidentPayload) (*Incident, error) {
	components, err := c.GetPaginatedComponents(
		NewAllPaginatedRequest(PaginatedRequestFilter{"status_page": input.StatusPage}),
	)
	if err != nil {
		return nil, err
	}

	affectedComponents, err := BuildAffectedComponents(input.Components, components.Results)
	if err != nil {
		return nil, err
	}

	incident := &Incident{
		StatusPage:   input.StatusPage,
		StartAt:      input.StartAt,
		Status:       IncidentStatusType(input.Status),
		Title:        input.Title,
		Description:  input.Description,
		Notify:       input.Notify,
		ShowOnTop:    input.ShowOnTop,
		AffectUptime: input.AffectUptime,
		PrivateNote:  input.PrivateNote,
		Components:   affectedComponents,
	}

	resp, err := c.Post("/api/incident/", incident)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create incident: %s\n%s", resp.Status, string(body))
	}

	var newIncident Incident
	err = parseResponseBody(resp, &newIncident)
	if err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &newIncident, nil
}

type PatchIncidentPayload struct {
	Title        *string             `json:"title,omitempty"`
	PrivateNote  *string             `json:"private_note,omitempty"`
	Notify       *bool               `json:"notify,omitempty"`
	ShowOnTop    *bool               `json:"show_on_top,omitempty"`
	AffectUptime *bool               `json:"affect_uptime,omitempty"`
	Components   []AffectedComponent `json:"components,omitempty"`
	EndAt        *time.Time          `json:"end_at,omitempty"`
}

func (c *Client) PatchIncident(uuid string, payload *PatchIncidentPayload) (*Incident, error) {
	resp, err := c.Patch(fmt.Sprintf("/api/incident/%s/", uuid), payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update incident: %s\n%s", resp.Status, string(body))
	}

	var incident Incident
	if err := parseResponseBody(resp, &incident); err != nil {
		return nil, err
	}
	return &incident, nil
}

func (c *Client) GetIncidentByUUID(uuid string) (*Incident, error) {
	resp, err := c.Get(fmt.Sprintf("/api/incident/%s/", uuid), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("incident %s: %s", uuid, resp.Status)
	}
	var inc Incident
	if err := parseResponseBody(resp, &inc); err != nil {
		return nil, err
	}
	return &inc, nil
}

func (c *Client) GetIncidentByID(id int) (*Incident, error) {
	result, err := c.GetPaginatedIncidents(NewAllPaginatedRequest(PaginatedRequestFilter{"id": id}))
	if err != nil {
		return nil, err
	}
	for i := range result.Results {
		if result.Results[i].ID != nil && *result.Results[i].ID == id {
			return &result.Results[i], nil
		}
	}
	return nil, fmt.Errorf("incident id=%d not found", id)
}

func (c *Client) GetPaginatedIncidents(payload PaginatedRequest) (*Paginated[Incident], error) {
	queryParams := ConvertToQueryParams(payload)

	resp, err := c.Get("/api/incident/", queryParams)

	if err != nil {
		return nil, errors.New("failed to perform GET request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve incidents: status " + resp.Status)
	}

	var result Paginated[Incident]
	err = parseResponseBody(resp, &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &result, nil
}
