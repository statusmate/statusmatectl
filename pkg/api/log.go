package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

type LogData struct {
	Incident    IncidentLogData    `json:"incident"`
	Maintenance MaintenanceLogData `json:"maintenance"`
	Update      UpdateLogData      `json:"update"`
}

type IncidentLogData struct {
	Title  string     `json:"title"`
	Impact ImpactType `json:"impact"`
}

type MaintenanceLogData struct {
	Title  string                `json:"title"`
	Status MaintenanceStatusType `json:"status"`
}

type UpdateLogData struct {
	Status      IncidentStatusType `json:"status"`
	Description string             `json:"description"`
}

const (
	LogEventsIncidentCreated    LogEventsEnum = "incident_created_incident"
	LogEventsMaintenanceCreated LogEventsEnum = "maintenance_created"
	LogEventsIncidentUpdated    LogEventsEnum = "incident_updated"
	LogEventsMaintenanceUpdated LogEventsEnum = "maintenance_updated"
	LogEventsUpdateCreated      LogEventsEnum = "update_created"
	LogEventsUpdateUpdated      LogEventsEnum = "update_updated"
	LogEventsUpdateDeleted      LogEventsEnum = "update_deleted"
)

type LogEventsEnum string

type Log struct {
	ID          int           `json:"id"`
	Event       LogEventsEnum `json:"event"`
	Actor       string        `json:"actor"`
	Data        LogData       `json:"data"`
	DataBefore  LogData       `json:"data_before"`
	UUID        string        `json:"uuid"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Incident    *int          `json:"incident"`
	Maintenance *int          `json:"maintenance"`
}

func (c *Client) GetPaginatedLogs(payload PaginatedRequest) (*Paginated[Log], error) {
	queryParams := ConvertToQueryParams(payload)

	resp, err := c.Get("/api/logs/", queryParams)
	if err != nil {
		return nil, errors.New("failed to perform GET request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve logs: status " + resp.Status)
	}

	var result Paginated[Log]
	if err := parseResponseBody(resp, &result); err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &result, nil
}
