package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Update представляет структуру обновления инцидента или обслуживания
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
