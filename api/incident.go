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

type Incident struct {
	ID           *int                `json:"id,omitempty"`
	UUID         *string             `json:"uuid,omitempty"`
	AbsoluteURL  *string             `json:"absolute_url,omitempty"`
	Components   []AffectedComponent `json:"components"`
	CreatedAt    *time.Time          `json:"created_at,omitempty"`
	UpdatedAt    *time.Time          `json:"updated_at,omitempty"`
	LastUpdateAt *time.Time          `json:"last_update_at,omitempty"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	EndAt        time.Time           `json:"end_at"`
	Logs         []Log               `json:"logs"`
	Notify       bool                `json:"notify"`
	Status       IncidentStatusType  `json:"status"`
	Updates      []Update            `json:"updates"`
	StartAt      time.Time           `json:"start_at"`
	StatusPage   int                 `json:"status_page"`
	PrivateNote  string              `json:"private_note"`
	ShowOnTop    bool                `json:"show_on_top"`
	AffectUptime bool                `json:"affect_uptime"`
}

func NewIncident(statusPage int) *Incident {
	return &Incident{
		StatusPage:   statusPage,
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
