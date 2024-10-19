package api

type ImpactType string

const (
	ImpactTypeOperational         ImpactType = "operational"
	ImpactTypeUnderMaintenance    ImpactType = "under_maintenance"
	ImpactTypeDegradedPerformance ImpactType = "degraded_performance"
	ImpactTypePartialOutage       ImpactType = "partial_outage"
	ImpactTypeMajorOutage         ImpactType = "major_outage"
)
