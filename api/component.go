package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Component struct {
	ID          *int       `json:"id,omitempty"`
	UUID        *string    `json:"uuid,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	Index       int        `json:"index"`
	Name        string     `json:"name"`
	Enabled     bool       `json:"enabled"`
	Description string     `json:"description"`
	Impact      ImpactType `json:"impact"`
	Histogram   bool       `json:"histogram"`
	StatusPage  int        `json:"status_page"`
	Uptime      float64    `json:"uptime"`
	UptimeDirty bool       `json:"uptime_dirty"`
	Collapse    bool       `json:"collapse"`
	Parent      *int       `json:"parent"`
	Private     bool       `json:"private"`
	StartDate   *time.Time `json:"start_date"`
}

type BatchUpdateComponent struct {
	ID    int `json:"id"`
	Index int `json:"index"`
}

func (c *Client) GetPaginatedComponents(payload PaginatedRequest) (*Paginated[Component], error) {
	queryParams := ConvertToQueryParams(payload)

	resp, err := c.Get("/api/component/", queryParams)

	if err != nil {
		return nil, errors.New("failed to perform GET request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve components: status " + resp.Status)
	}

	var components Paginated[Component]
	err = parseResponseBody(resp, &components)
	if err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &components, nil
}

func (c *Client) GetPaginatedComponentsByStatusPageSlug(statusPageSlug string) (*Paginated[Component], error) {
	request := NewAllPaginatedRequest(map[string]interface{}{
		"status_page": statusPageSlug,
	})
	return c.GetPaginatedComponents(request)
}

func (c *Client) CreateComponent(component *Component) (*Component, error) {
	resp, err := c.Post("/api/component/", component)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New("failed to create component: status " + resp.Status)
	}

	var newComponent Component
	err = parseResponseBody(resp, &newComponent)
	if err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &newComponent, nil
}

// GetComponentByUUID получает компонент по его UUID.
func (c *Client) GetComponentByUUID(uuid string) (*Component, error) {
	// Формируем URL с параметром uuid
	url := fmt.Sprintf("/api/component/%s/", uuid)

	resp, err := c.Get(url, nil)
	if err != nil {
		return nil, errors.New("failed to perform GET request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve component: status " + resp.Status)
	}

	var component Component
	if err := parseResponseBody(resp, &component); err != nil {
		return nil, errors.New("failed to parse response body: " + err.Error())
	}

	return &component, nil
}

func (c *Client) UpdateComponent(uuid string, partial *Component) (*Component, error) {
	// Формируем URL с параметром uuid
	url := fmt.Sprintf("/api/component/%s/", uuid)

	resp, err := c.Patch(url, partial)
	if err != nil {
		return nil, errors.New("failed to perform PATCH request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to update component: status " + resp.Status)
	}

	var updatedComponent Component
	if err := parseResponseBody(resp, &updatedComponent); err != nil {
		return nil, errors.New("failed to parse response body: " + err.Error())
	}

	return &updatedComponent, nil
}

func (c *Client) BatchUpdateComponent(partial []BatchUpdateComponent) (*Component, error) {
	resp, err := c.Post("/api/component/batch_update/", partial)
	if err != nil {
		return nil, errors.New("failed to perform PATCH request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to update component: status " + resp.Status)
	}

	var updatedComponent Component
	if err := parseResponseBody(resp, &updatedComponent); err != nil {
		return nil, errors.New("failed to parse response body: " + err.Error())
	}

	return &updatedComponent, nil
}

// DeleteComponent удаляет компонент по его UUID.
func (c *Client) DeleteComponent(uuid string) error {
	// Формируем URL с параметром uuid
	url := fmt.Sprintf("/api/component/%s/", uuid)

	resp, err := c.Delete(url)
	if err != nil {
		return errors.New("failed to perform DELETE request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return errors.New("failed to delete component: status " + resp.Status)
	}

	return nil
}
