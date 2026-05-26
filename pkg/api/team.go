package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TeamUser struct {
	ID        int        `json:"id"`
	UUID      string     `json:"uuid" tab:"UUID"`
	Role      string     `json:"role" tab:"Role"`
	User      int        `json:"user" tab:"UserID"`
	Team      int        `json:"team"`
	IsActive  bool       `json:"is_active" tab:"Active"`
	CreatedAt *time.Time `json:"created_at,omitempty" tab:"Created"`
}

type TeamUserNested struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type TeamUserExpanded struct {
	ID        int            `json:"id"`
	UUID      string         `json:"uuid"`
	Role      string         `json:"role"`
	User      TeamUserNested `json:"user"`
	Team      int            `json:"team"`
	IsActive  bool           `json:"is_active"`
	CreatedAt *time.Time     `json:"created_at,omitempty"`
}

type Team struct {
	ID                int        `json:"id"`
	UUID              string     `json:"uuid"`
	Name              string     `json:"name"`
	Slug              string     `json:"slug"`
	Description       *string    `json:"description,omitempty"`
	TeamUser          *TeamUser  `json:"team_user,omitempty"`
	PremiumPagesCount int        `json:"premium_pages_count"`
	CreatedAt         *time.Time `json:"created_at,omitempty"`
	UpdatedAt         *time.Time `json:"updated_at,omitempty"`
}

type TeamInvite struct {
	ID         int        `json:"id"`
	UUID       string     `json:"uuid"`
	Email      string     `json:"email"`
	Code       string     `json:"code"`
	Role       string     `json:"role"`
	Team       int        `json:"team"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	CreatedBy  *int       `json:"created_by,omitempty"`
	AcceptedBy *int       `json:"accepted_by,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

type CreateTeamInvitePayload struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	Team  int    `json:"team"`
}

func (c *Client) GetPaginatedTeams(payload PaginatedRequest) (*Paginated[Team], error) {
	queryParams := ConvertToQueryParams(payload)

	resp, err := c.Get("/api/teams/", queryParams)
	if err != nil {
		return nil, errors.New("failed to perform GET request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve teams: status " + resp.Status)
	}

	var result Paginated[Team]
	if err := parseResponseBody(resp, &result); err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &result, nil
}

func (c *Client) GetPaginatedTeamInvites(payload PaginatedRequest) (*Paginated[TeamInvite], error) {
	queryParams := ConvertToQueryParams(payload)

	resp, err := c.Get("/api/team_invite/", queryParams)
	if err != nil {
		return nil, errors.New("failed to perform GET request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve team invites: status " + resp.Status)
	}

	var result Paginated[TeamInvite]
	if err := parseResponseBody(resp, &result); err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &result, nil
}

func (c *Client) CreateTeamInvite(input *CreateTeamInvitePayload) (*TeamInvite, error) {
	resp, err := c.Post("/api/team_invite/", input)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create team invite: %s\n%s", resp.Status, string(body))
	}

	var invite TeamInvite
	if err := parseResponseBody(resp, &invite); err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &invite, nil
}

func (c *Client) GetPaginatedTeamUsers(payload PaginatedRequest) (*Paginated[TeamUser], error) {
	queryParams := ConvertToQueryParams(payload)

	resp, err := c.Get("/api/team_user/", queryParams)
	if err != nil {
		return nil, errors.New("failed to perform GET request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve team users: status " + resp.Status)
	}

	var result Paginated[TeamUser]
	if err := parseResponseBody(resp, &result); err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &result, nil
}

func (c *Client) GetPaginatedTeamUsersExpanded(payload PaginatedRequest) (*Paginated[TeamUserExpanded], error) {
	queryParams := ConvertToQueryParams(payload)
	queryParams["expand"] = "user"

	resp, err := c.Get("/api/team_user/", queryParams)
	if err != nil {
		return nil, errors.New("failed to perform GET request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve team users: status " + resp.Status)
	}

	var result Paginated[TeamUserExpanded]
	if err := parseResponseBody(resp, &result); err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &result, nil
}

func (c *Client) DeleteTeamInvite(code string) error {
	resp, err := c.Delete(fmt.Sprintf("/api/team_invite/%s/", code))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete team invite: %s\n%s", resp.Status, string(body))
	}

	return nil
}
