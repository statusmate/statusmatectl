package api

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
