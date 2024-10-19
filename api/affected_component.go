package api

import "time"

type AffectedComponent struct {
	Component int        `json:"component"`
	Impact    ImpactType `json:"impact"`
	ID        *int       `json:"id,omitempty"`
	UUID      *string    `json:"uuid,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}
