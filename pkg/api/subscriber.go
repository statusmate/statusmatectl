package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Subscriber struct {
	UUID               *string    `json:"uuid,omitempty" tab:"UUID"`
	Email              string     `json:"email" tab:"Email"`
	StatusPage         *int       `json:"status_page,omitempty"`
	ReleasePage        *int       `json:"release_page,omitempty"`
	SubscribeByEmail   bool       `json:"subscribe_by_email" tab:"Email"`
	SubscribeByWebhook bool       `json:"subscribe_by_webhook" tab:"Webhook"`
	HasPassword        bool       `json:"has_password" tab:"HasPwd"`
	Confirmed           bool      `json:"confirmed" tab:"Confirmed"`
	WebhookURL         string     `json:"webhook_url,omitempty"`
	SubscribeTo		   string 	  `json:"subscribe_to" tab:"Subscribe to"`
	Components 	 	   []int 	  `json:"components"`
	CreatedAt          *time.Time `json:"created_at,omitempty" tab:"Created"`
}

type CreateSubscriberPayload struct {
	Email            string `json:"email"`
	StatusPage       int    `json:"status_page"`
	SubscribeByEmail bool   `json:"subscribe_by_email"`
}

func (c *Client) GetPaginatedSubscribers(payload PaginatedRequest) (*Paginated[Subscriber], error) {
	queryParams := ConvertToQueryParams(payload)

	resp, err := c.Get("/api/subscriber/", queryParams)
	if err != nil {
		return nil, errors.New("failed to perform GET request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve subscribers: status " + resp.Status)
	}

	var result Paginated[Subscriber]
	if err := parseResponseBody(resp, &result); err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &result, nil
}

func (c *Client) CreateSubscriber(input *CreateSubscriberPayload) (*Subscriber, error) {
	resp, err := c.Post("/api/subscriber/", input)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create subscriber: %s\n%s", resp.Status, string(body))
	}

	var sub Subscriber
	if err := parseResponseBody(resp, &sub); err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &sub, nil
}

func (c *Client) DeleteSubscriber(uuid string) error {
	resp, err := c.Delete(fmt.Sprintf("/api/subscriber/%s/", uuid))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete subscriber: %s\n%s", resp.Status, string(body))
	}

	return nil
}

func (c *Client) VerifySubscriber(uuid string) error {
	payload := map[string]any{"confirmed": true}
	resp, err := c.Patch(fmt.Sprintf("/api/subscriber/%s/", uuid), payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to verify subscriber: %s\n%s", resp.Status, string(body))
	}

	return nil
}
