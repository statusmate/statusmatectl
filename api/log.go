package api

import "time"

// LogData представляет структуру данных для лога
type LogData struct {
	Incident    IncidentLogData    `json:"incident"`
	Maintenance MaintenanceLogData `json:"maintenance"`
	Update      UpdateLogData      `json:"update"`
}

// IncidentLogData представляет структуру данных инцидента в логе
type IncidentLogData struct {
	Title  string     `json:"title"`
	Impact ImpactType `json:"impact"`
}

// MaintenanceLogData представляет структуру данных техобслуживания в логе
type MaintenanceLogData struct {
	Title  string                `json:"title"`
	Status MaintenanceStatusType `json:"status"`
}

// UpdateLogData представляет структуру данных обновления в логе
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

// LogEventsEnum представляет тип событий лога
type LogEventsEnum string

// Log представляет структуру лога
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

// Actor представляет информацию об актере, который инициировал событие
type Actor struct {
	Name string `json:"name"`
	Type string `json:"type"` // "user" или "system"
}
