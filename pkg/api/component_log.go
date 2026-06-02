package api

import (
	"sort"
	"time"
)

type ComponentLogEntry struct {
	At       time.Time
	UUID     string
	Object   string // "incident" | "maintenance"
	Title    string
	Status   string
	Desc     string
	ParentID int
}

// GetComponentLogEntries returns all update entries for incidents and maintenances
// affecting the given component. eventType filters by "incident", "maintenance", or "" for both.
func (c *Client) GetComponentLogEntries(componentID, statusPageID int, eventType string) ([]ComponentLogEntry, error) {
	pageFilter := PaginatedRequestFilter{
		"status_page": statusPageID,
		"components":  []int{componentID},
	}

	var entries []ComponentLogEntry

	if eventType == "" || eventType == "incident" {
		incidents, err := c.GetPaginatedIncidents(NewAllPaginatedRequest(pageFilter))
		if err != nil {
			return nil, err
		}
		for _, inc := range incidents.Results {
			if inc.ID == nil {
				continue
			}
			updates, err := c.GetPaginatedUpdates(
				NewAllPaginatedRequest(PaginatedRequestFilter{"incident": *inc.ID}),
			)
			if err != nil {
				continue
			}
			for _, u := range updates.Results {
				entries = append(entries, ComponentLogEntry{
					At:       u.At,
					Object:   "incident",
					UUID:     u.UUID,
					Title:    inc.Title,
					Status:   u.Status,
					Desc:     u.Description,
					ParentID: *inc.ID,
				})
			}
		}
	}

	if eventType == "" || eventType == "maintenance" {
		maintenances, err := c.GetPaginatedMaintenance(NewAllPaginatedRequest(pageFilter))
		if err != nil {
			return nil, err
		}
		for _, m := range maintenances.Results {
			if m.ID == nil {
				continue
			}
			updates, err := c.GetPaginatedUpdates(
				NewAllPaginatedRequest(PaginatedRequestFilter{"maintenance": *m.ID}),
			)
			if err != nil {
				continue
			}
			for _, u := range updates.Results {
				entries = append(entries, ComponentLogEntry{
					At:       u.At,
					Object:   "maintenance",
					UUID:     u.UUID,
					Title:    m.Title,
					Status:   u.Status,
					Desc:     u.Description,
					ParentID: *m.ID,
				})
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].At.After(entries[j].At)
	})

	return entries, nil
}
