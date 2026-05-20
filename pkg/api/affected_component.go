package api

import (
	"time"
)

type AffectedComponent struct {
	ID        *int       `json:"id,omitempty"`
	UUID      *string    `json:"uuid,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	Component int        `json:"component"`
	Impact    ImpactType `json:"impact"`
}

func NewAffectedComponent(component int, impact ImpactType) *AffectedComponent {
	return &AffectedComponent{
		Component: component,
		Impact:    impact,
	}
}
