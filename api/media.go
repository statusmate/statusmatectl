package api

import "time"

type Media struct {
	ID        int       `json:"id"`
	Path      string    `json:"path"`
	UUID      string    `json:"uuid"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	FileName  string    `json:"file_name"`
	Team      int       `json:"team"`
}
