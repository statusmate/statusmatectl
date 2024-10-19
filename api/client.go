package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	BaseURL   string
	AuthToken string
	Username  string
	Email     string
	client    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		client:  &http.Client{
			//Transport: &loggingTransport{},
		},
	}
}

// NewClientWithToken читает токен из файла и создает клиента
func NewClientWithToken(baseURL string) (*Client, error) {
	client := NewClient(baseURL)

	authRC, err := client.LoadAuthRC(baseURL)
	if err != nil {
		return nil, errors.New("need auth You need to authorize this machine using `statusmate login`")
	} else {
		client.SetEmail(authRC.Email)
		client.SetUsername(authRC.Username)
		client.SetAuthToken(authRC.Token)
	}

	return client, nil
}

// SetAuthToken устанавливает токен авторизации
func (c *Client) SetAuthToken(token string) {
	c.AuthToken = token
}

func (c *Client) SetUsername(username string) {
	c.Username = username
}

func (c *Client) SetEmail(email string) {
	c.Email = email
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	if c.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.AuthToken))
	}
	req.Header.Set("Content-Type", "application/json")
	return c.client.Do(req)
}

// Get выполняет GET-запрос
func (c *Client) Get(endpoint string, queryParams map[string]string) (*http.Response, error) {
	// Создаем новый URL с базовым URL и конечной точкой
	fullURL, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return nil, err
	}

	// Создаем URL-кодировщик
	q := fullURL.Query()

	// Добавляем query параметры
	for key, value := range queryParams {
		q.Add(key, value)
	}

	// Устанавливаем обновленные query параметры в URL
	fullURL.RawQuery = q.Encode()

	// Создаем новый HTTP-запрос
	req, err := http.NewRequest(http.MethodGet, fullURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return c.doRequest(req)
}

// Post выполняет POST-запрос с телом запроса
func (c *Client) Post(endpoint string, body interface{}) (*http.Response, error) {
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
func (c *Client) Patch(endpoint string, body interface{}) (*http.Response, error) {
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
func parseResponseBody(resp *http.Response, v interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}
