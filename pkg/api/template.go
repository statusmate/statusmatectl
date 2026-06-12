package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

type TagShort struct {
	ID    int    `json:"id"`
	UUID  string `json:"uuid"`
	Title string `json:"title"`
	Color string `json:"color,omitempty"`
}

type Template struct {
	ID           *int                `json:"id,omitempty"`
	UUID         *string             `json:"uuid,omitempty"`
	Title        string              `json:"title"`
	FriendlyName string              `json:"friendly_name,omitempty"`
	Description  string              `json:"description,omitempty"`
	Notify       bool                `json:"notify"`
	Status       *string             `json:"status,omitempty"`
	StatusPage   int                 `json:"status_page"`
	Components   []AffectedComponent `json:"components"`
	AssignedTags []TagShort          `json:"assigned_tags,omitempty"`
	CreatedAt    *time.Time          `json:"created_at,omitempty"`
	UpdatedAt    *time.Time          `json:"updated_at,omitempty"`
}

func (c *Client) GetPaginatedTemplates(payload PaginatedRequest) (*Paginated[Template], error) {
	queryParams := ConvertToQueryParams(payload)
	resp, err := c.Get("/api/template/", queryParams)
	if err != nil {
		return nil, errors.New("failed to perform GET request: " + err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve templates: status " + resp.Status)
	}
	var result Paginated[Template]
	if err := parseResponseBody(resp, &result); err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}
	return &result, nil
}

func (c *Client) GetTemplate(uuid string) (*Template, error) {
	resp, err := c.Get(fmt.Sprintf("/api/template/%s/", uuid), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("template %s: %s", uuid, resp.Status)
	}
	var t Template
	if err := parseResponseBody(resp, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *Client) DeleteTemplate(uuid string) error {
	resp, err := c.Delete(fmt.Sprintf("/api/template/%s/", uuid))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		body := make([]byte, 256)
		n, _ := resp.Body.Read(body)
		return fmt.Errorf("failed to delete template: %s\n%s", resp.Status, body[:n])
	}
	return nil
}
