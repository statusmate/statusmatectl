package api

import (
	"errors"
	"fmt"
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
	UUID         *string             `json:"uuid,omitempty" tab:"UUID"`
	AbsoluteURL  *string             `json:"absolute_url,omitempty" tab:"URL"`
	Title        string              `json:"title" tab:"Title"`
	Status       IncidentStatusType  `json:"status" tab:"Status"`
	ID           *int                `json:"id,omitempty"`
	Components   []AffectedComponent `json:"components"`
	CreatedAt    *time.Time          `json:"created_at,omitempty"`
	CreatedBy    *int                `json:"created_by,omitempty"`
	UpdatedAt    *time.Time          `json:"updated_at,omitempty"`
	LastUpdateAt *time.Time          `json:"last_update_at,omitempty"`
	Description  string              `json:"description"`
	EndAt        time.Time           `json:"end_at"`
	Logs         []Log               `json:"logs"`
	Notify       bool                `json:"notify"`
	Updates      []IncidentUpdate    `json:"updates"`
	StartAt      *time.Time          `json:"start_at"`
	StatusPage   int                 `json:"status_page"`
	PrivateNote  string              `json:"private_note"`
	ShowOnTop    bool                `json:"show_on_top"`
	AffectUptime bool                `json:"affect_uptime"`
}

func NewIncident(statusPage *StatusPage) (*Incident, error) {
	startAt := time.Now()

	return &Incident{
		StatusPage:   statusPage.ID,
		Status:       IncidentStatusInvestigation,
		Components:   []AffectedComponent{},
		Title:        "",
		Description:  "",
		PrivateNote:  "",
		Notify:       true,
		StartAt:      &startAt,
		ShowOnTop:    true,
		AffectUptime: true,
	}, nil
}

func (c *Client) CreateIncident(incident *Incident) (*Incident, error) {
	resp, err := c.Post("/api/incident/", incident)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New("failed to create incident: status " + resp.Status)
	}

	var newIncident Incident
	err = parseResponseBody(resp, &newIncident)
	if err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &newIncident, nil
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
