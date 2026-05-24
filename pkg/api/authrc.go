package api

type AuthRC struct {
	API                string `json:"api"`
	Token              string `json:"token"`
	DefaultStatusPage  string `json:"default_status_page"`
	DefaultReleasePage string `json:"default_release_page"`
	DefaultTeam        int    `json:"default_team,omitempty"`
}
