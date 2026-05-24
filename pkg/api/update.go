package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Update[T any] struct {
	ID          int                 `json:"id"`
	At          time.Time           `json:"at"`
	Components  []AffectedComponent `json:"components"`
	Description string              `json:"description"`
	Notify      bool                `json:"notify"`
	Status      T                   `json:"status"`
	UUID        string              `json:"uuid"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Incident    *int                `json:"incident,omitempty"`
	Maintenance *int                `json:"maintenance,omitempty"`
}

type MaintenanceUpdate Update[MaintenanceStatusType]
type IncidentUpdate Update[IncidentStatusType]

func NewUpdateForIncident(incident *Incident) (*IncidentUpdate, error) {
	if incident.ID == nil {
		return nil, errors.New("incident must field id")
	}

	var update = &IncidentUpdate{
		Incident:    incident.ID,
		Notify:      true,
		Status:      IncidentStatusIdentified,
		Description: "",
		At:          time.Now(),
		Components:  make([]AffectedComponent, 0),
	}

	return update, nil
}

func NewUpdateForMaintenance(maintenance *Maintenance) (*MaintenanceUpdate, error) {
	if maintenance.ID == nil {
		return nil, errors.New("maintenance must field id")
	}

	var update = &MaintenanceUpdate{
		Maintenance: maintenance.ID,
		Notify:      true,
		Status:      MaintenanceStatusInProgress,
		Description: "",
		At:          time.Now(),
		Components:  make([]AffectedComponent, 0),
	}

	return update, nil
}

func (c *Client) CreateUpdate(update *Update[any]) (*Update[any], error) {
	resp, err := c.Post("/api/update/", update)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New("failed to create update: status " + resp.Status)
	}

	var newUpdate Update[any]
	err = parseResponseBody(resp, &newUpdate)
	if err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &newUpdate, nil
}

func (c *Client) GetPaginatedUpdates(payload PaginatedRequest) (*Paginated[Update[string]], error) {
	queryParams := ConvertToQueryParams(payload)

	resp, err := c.Get("/api/update/", queryParams)
	if err != nil {
		return nil, errors.New("failed to perform GET request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve updates: status " + resp.Status)
	}

	var result Paginated[Update[string]]
	if err := parseResponseBody(resp, &result); err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &result, nil
}

func (c *Client) GetUpdateByUUID(uuid string) (*Update[string], error) {
	result, err := c.GetPaginatedUpdates(NewAllPaginatedRequest(PaginatedRequestFilter{"uuid": uuid}))
	if err != nil {
		return nil, err
	}
	if len(result.Results) == 0 {
		return nil, fmt.Errorf("update %s not found", uuid)
	}
	return &result.Results[0], nil
}

func (c *Client) CreateMaintenanceUpdate(update *MaintenanceUpdate) error {
	resp, err := c.Post("/api/update/", update)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create maintenance update: %s\n%s", resp.Status, string(body))
	}

	return nil
}

type CreateIncidentUpdatePayload struct {
	Description string   `format:"description"`
	Status      string   `format:"status"`
	Components  []string `format:"components"`
	Notify      bool     `format:"notify"`
}

var CreateIncidentUpdatePayloadFieldDescriptions = map[string]string{
	"description": "Сообщение обновления",
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
	"notify": "Отправлять ли уведомление пользователям",
}

func (c *Client) GetLatestIncidentUpdate(incidentID int) (*Update[string], error) {
	result, err := c.GetPaginatedUpdates(NewAllPaginatedRequest(PaginatedRequestFilter{"incident": incidentID}))
	if err != nil {
		return nil, err
	}
	if len(result.Results) == 0 {
		return nil, nil
	}
	latest := &result.Results[0]
	for i := range result.Results[1:] {
		u := &result.Results[i+1]
		if u.At.After(latest.At) {
			latest = u
		}
	}
	return latest, nil
}

func (c *Client) CreateIncidentUpdate(update *IncidentUpdate) error {
	resp, err := c.Post("/api/update/", update)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create incident update: %s\n%s", resp.Status, string(body))
	}

	return nil
}
