package api

import "time"

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
	ID         int           `json:"id"`
	Event      LogEventsEnum `json:"event"`
	Actor      Actor         `json:"actor"`
	Data       LogData       `json:"data"`
	DataBefore LogData       `json:"data_before"`
	UUID       string        `json:"uuid"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

type Actor struct {
	Name string `json:"name"`
	Type string `json:"type"` // "user" или "system"
}
