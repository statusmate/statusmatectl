package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Update представляет структуру обновления инцидента или обслуживания
type Update struct {
	ID          int                 `json:"id"`
	At          time.Time           `json:"at"`
	Components  []AffectedComponent `json:"components"`
	Description string              `json:"description"`
	Notify      bool                `json:"notify"`
	Status      IncidentStatusType  `json:"status"`
	UUID        string              `json:"uuid"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Incident    *int                `json:"incident,omitempty"`
	Maintenance *int                `json:"maintenance,omitempty"`
}

func (c *Client) CreateUpdate(update *Update) (*Update, error) {
	resp, err := c.Post("/api/update/", update)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New("failed to create update: status " + resp.Status)
	}

	var newUpdate Update
	err = parseResponseBody(resp, &newUpdate)
	if err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &newUpdate, nil
}
