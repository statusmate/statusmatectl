package api

import (
	"errors"
	"fmt"
	"strings"
)

type ImpactType string

const (
	ImpactTypeOperational         ImpactType = "operational"
	ImpactTypeUnderMaintenance    ImpactType = "under_maintenance"
	ImpactTypeDegradedPerformance ImpactType = "degraded_performance"
	ImpactTypePartialOutage       ImpactType = "partial_outage"
	ImpactTypeMajorOutage         ImpactType = "major_outage"
)

var impactMap = map[string]ImpactType{
	"o":           ImpactTypeOperational,
	"op":          ImpactTypeOperational,
	"operational": ImpactTypeOperational,

	"u":                 ImpactTypeUnderMaintenance,
	"um":                ImpactTypeUnderMaintenance,
	"under_maintenance": ImpactTypeUnderMaintenance,

	"d":                    ImpactTypeDegradedPerformance,
	"dp":                   ImpactTypeDegradedPerformance,
	"degraded_performance": ImpactTypeDegradedPerformance,

	"p":              ImpactTypePartialOutage,
	"po":             ImpactTypePartialOutage,
	"partial_outage": ImpactTypePartialOutage,

	"m":            ImpactTypeMajorOutage,
	"mo":           ImpactTypeMajorOutage,
	"major_outage": ImpactTypeMajorOutage,
}

type ComponentImpact struct {
	Component string
	Impact    ImpactType
}

func ParseComponentImpact(input string) (ComponentImpact, error) {
	parts := strings.SplitN(input, " ", 2)
	if len(parts) != 2 {
		return ComponentImpact{}, errors.New("invalid format, expected 'impact component'")
	}

	impactStr := parts[0]
	component := parts[1]

	impact, exists := impactMap[impactStr]
	if !exists {
		return ComponentImpact{}, fmt.Errorf("invalid impact type: %s", impactStr)
	}

	return ComponentImpact{Component: component, Impact: impact}, nil
}
