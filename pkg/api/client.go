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
	AuthRC  *AuthRC
	http    *http.Client
	logger  *slog.Logger
}

type QueryParams = map[string]any

func NewClient(baseURL string, logger *slog.Logger) *Client {
	httpClient := &http.Client{
		Transport: &loggingTransport{
			Logger:    logger,
			Transport: http.DefaultTransport,
		},
	}

	return &Client{
		logger:  logger,
		BaseURL: fmt.Sprintf("https://%s", baseURL),
		http:    httpClient,
	}
}

func (c *Client) SetAuthToken(token string) {
	if c.AuthRC == nil {
		c.AuthRC = &AuthRC{}
	}
	c.AuthRC.Token = token
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	if c.AuthRC != nil && c.AuthRC.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.AuthRC.Token))
	}
	req.Header.Set("Content-Type", "application/json")
	return c.http.Do(req)
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

func (c *Client) Delete(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	return c.doRequest(req)
}

func parseResponseBody(resp *http.Response, v any) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}
