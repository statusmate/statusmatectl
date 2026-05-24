package api

import (
	"errors"
	"fmt"
	"io"
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

type CreateMaintenancePayload struct {
	Title        string   `format:"title"`
	Description  string   `format:"description"`
	StartAt      string   `format:"start_at"`
	EndAt        string   `format:"end_at"`
	Components   []string `format:"components"`
	Notify       bool     `format:"notify"`
	AutoStart    bool     `format:"auto_start"`
	AutoEnd      bool     `format:"auto_end"`
	AffectUptime bool     `format:"affect_uptime"`
	StatusPage   int
}

var CreateMaintenancePayloadFieldDescriptions = map[string]string{
	"title":       "Название обслуживания",
	"description": "Описание обслуживания",
	"start_at":    "Время начала (RFC3339, например 2024-01-01T10:00:00+03:00)",
	"end_at":      "Время окончания (RFC3339, опционально — оставьте пустым)",
	"components": `Указываются в формате: [Impact] [Имя компонента]
Возможные значения Impact:
- u, um  — under maintenance (рекомендуется)
- o, op  — operational
- d, dp  — degraded performance
- p, po  — partial outage
- m, mo  — major outage

Примеры:
u Web
u API`,
	"notify":        "Отправлять ли уведомление пользователям (yes/no)",
	"auto_start":    "Автоматически начать обслуживание в start_at (yes/no)",
	"auto_end":      "Автоматически завершить обслуживание в end_at (yes/no)",
	"affect_uptime": "Влияет ли обслуживание на аптайм компонента (yes/no)",
}

func NewCreateMaintenancePayload(statusPage *StatusPage) *CreateMaintenancePayload {
	return &CreateMaintenancePayload{
		StatusPage:   statusPage.ID,
		Title:        "",
		Description:  "",
		StartAt:      time.Now().Format(time.RFC3339),
		EndAt:        "",
		Components:   []string{},
		Notify:       true,
		AutoStart:    false,
		AutoEnd:      false,
		AffectUptime: true,
	}
}

type maintenanceCreateRequest struct {
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	StatusPage   int                 `json:"status_page"`
	StartAt      time.Time           `json:"start_at"`
	EndAt        *time.Time          `json:"end_at,omitempty"`
	Notify       bool                `json:"notify"`
	AutoStart    bool                `json:"auto_start"`
	AutoEnd      bool                `json:"auto_end"`
	AffectUptime bool                `json:"affect_uptime"`
	ShowOnPage   bool                `json:"show_on_page"`
	Components   []AffectedComponent `json:"components"`
}

func (c *Client) CreateMaintenance(input *CreateMaintenancePayload) (*Maintenance, error) {
	comps, err := c.GetPaginatedComponents(
		NewAllPaginatedRequest(PaginatedRequestFilter{"status_page": input.StatusPage}),
	)
	if err != nil {
		return nil, err
	}

	affectedComponents, err := BuildAffectedComponents(input.Components, comps.Results)
	if err != nil {
		return nil, err
	}

	startAt, err := time.Parse(time.RFC3339, input.StartAt)
	if err != nil {
		return nil, fmt.Errorf("invalid start_at %q: %v", input.StartAt, err)
	}

	var endAt *time.Time
	if input.EndAt != "" {
		t, err := time.Parse(time.RFC3339, input.EndAt)
		if err != nil {
			return nil, fmt.Errorf("invalid end_at %q: %v", input.EndAt, err)
		}
		endAt = &t
	}

	req := &maintenanceCreateRequest{
		Title:        input.Title,
		Description:  input.Description,
		StatusPage:   input.StatusPage,
		StartAt:      startAt,
		EndAt:        endAt,
		Notify:       input.Notify,
		AutoStart:    input.AutoStart,
		AutoEnd:      input.AutoEnd,
		AffectUptime: input.AffectUptime,
		ShowOnPage:   true,
		Components:   affectedComponents,
	}

	resp, err := c.Post("/api/maintenance/", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create maintenance: %s\n%s", resp.Status, string(body))
	}

	var m Maintenance
	if err := parseResponseBody(resp, &m); err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &m, nil
}

func (c *Client) GetMaintenanceByUUID(uuid string) (*Maintenance, error) {
	resp, err := c.Get(fmt.Sprintf("/api/maintenance/%s/", uuid), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("maintenance %s: %s", uuid, resp.Status)
	}
	var m Maintenance
	if err := parseResponseBody(resp, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *Client) GetMaintenanceByID(id int) (*Maintenance, error) {
	result, err := c.GetPaginatedMaintenance(NewAllPaginatedRequest(PaginatedRequestFilter{"id": id}))
	if err != nil {
		return nil, err
	}
	for i := range result.Results {
		if result.Results[i].ID != nil && *result.Results[i].ID == id {
			return &result.Results[i], nil
		}
	}
	return nil, fmt.Errorf("maintenance id=%d not found", id)
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
