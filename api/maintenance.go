package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

type MaintenanceStatusType string

const (
	MaintenanceStatusNotStarted MaintenanceStatusType = "maintenance_not_started"
	MaintenanceStatusInProgress MaintenanceStatusType = "maintenance_in_progress"
	MaintenanceStatusCompleted  MaintenanceStatusType = "maintenance_completed"
	MaintenanceStatusNotice     MaintenanceStatusType = "notice"
)

func MaintenanceStatusList() []MaintenanceStatusType {
	return []MaintenanceStatusType{
		MaintenanceStatusNotStarted,
		MaintenanceStatusInProgress,
		MaintenanceStatusCompleted,
		MaintenanceStatusNotice,
	}
}

func MaintenanceActiveStatusList() []MaintenanceStatusType {
	return []MaintenanceStatusType{
		MaintenanceStatusNotStarted,
		MaintenanceStatusInProgress,
	}
}

type Maintenance struct {
	UUID                *string               `json:"uuid" tab:"UUID"`
	Title               string                `json:"title"  tab:"Title"`
	AbsoluteURL         string                `json:"absolute_url" tab:"URL"`
	StartAt             *time.Time            `json:"start_at" tab:"StartAt"`
	EndAt               *time.Time            `json:"end_at" tab:"EndAt"`
	Status              MaintenanceStatusType `json:"status" tab:"Status"`
	ID                  *int                  `json:"id"`
	Components          []AffectedComponent   `json:"components"`
	CreatedAt           *time.Time            `json:"created_at"`
	UpdatedAt           *time.Time            `json:"updated_at"`
	Description         string                `json:"description"`
	NotifyBeforeAt      *time.Time            `json:"notify_before_at"`
	Logs                []Log                 `json:"logs"`
	Notify              bool                  `json:"notify"`
	NotifyBefore        bool                  `json:"notify_before"`
	NotifyBeforeMinutes int                   `json:"notify_before_minutes"`
	AutoStart           bool                  `json:"auto_start"`
	NotifyAutoStart     bool                  `json:"notify_auto_start"`
	AutoEnd             bool                  `json:"auto_end"`
	NotifyAutoEnd       bool                  `json:"notify_auto_end"`
	LastUpdateAt        *time.Time            `json:"last_update_at"`
	StatusPage          int                   `json:"status_page"`
	PrivateNote         string                `json:"private_note"`
	Updates             []MaintenanceUpdate   `json:"updates"`
	AffectUptime        bool                  `json:"affect_uptime"`
}

func (c *Client) GetPaginatedMaintenance(payload PaginatedRequest) (*Paginated[Maintenance], error) {
	queryParams := ConvertToQueryParams(payload)

	resp, err := c.Get("/api/maintenance/", queryParams)

	if err != nil {
		return nil, errors.New("failed to perform GET request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve maintenance: status " + resp.Status)
	}

	var result Paginated[Maintenance]
	err = parseResponseBody(resp, &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &result, nil
}
