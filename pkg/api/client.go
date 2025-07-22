package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type Client struct {
	BaseURL string
	Token   string
	client  *http.Client
}

type QueryParams = map[string]interface{}

func NewClient(baseURL string, logger *slog.Logger) *Client {
	httpClient := &http.Client{
		Transport: &loggingTransport{
			Logger:    logger,
			Transport: http.DefaultTransport,
		},
	}

	return &Client{
		BaseURL: baseURL,
		client:  httpClient,
	}
}

func (c *Client) SetAuthToken(token string) {
	c.Token = token
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	if c.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.Token))
	}
	req.Header.Set("Content-Type", "application/json")
	return c.client.Do(req)
}

func (c *Client) Get(endpoint string, queryParams QueryParams) (*http.Response, error) {
	fullURL, err := c.stringifyQueryParams(endpoint, queryParams)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, fullURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return c.doRequest(req)
}

func (c *Client) Post(endpoint string, body any) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req)
}

// Patch выполняет PATCH-запрос с телом запроса
func (c *Client) Patch(endpoint string, body any) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPatch, c.BaseURL+endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req)
}

// Delete выполняет DELETE-запрос
func (c *Client) Delete(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	return c.doRequest(req)
}

// Utility method to parse the response body
func parseResponseBody(resp *http.Response, v any) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}
