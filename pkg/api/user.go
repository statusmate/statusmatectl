package api

import (
	"errors"
	"log"
	"net/http"
)

type User struct {
	LastLogin             string  `json:"last_login"`
	Username              string  `json:"username"`
	FirstName             string  `json:"first_name"`
	LastName              string  `json:"last_name"`
	Email                 string  `json:"email"`
	IsStaff               bool    `json:"is_staff"`
	IsActive              bool    `json:"is_active"`
	DateJoined            string  `json:"date_joined"`
	Name                  string  `json:"name"`
	Phone                 *string `json:"phone"` // pointer to handle null
	HasActiveSubscription bool    `json:"has_active_subscription"`
	EmailConfirmedAt      *string `json:"email_confirmed_at"` // pointer to handle null
	Groups                []int   `json:"groups"`
	UserPermissions       []int   `json:"user_permissions"`
	ReceiveNotifications  bool    `json:"receive_notifications"`
}

func (c *Client) GetMe() (*User, error) {
	resp, err := c.Get("/api/auth/me/", nil)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("authorization failed: invalid credentials")
	}

	var meResponse User
	err = parseResponseBody(resp, &meResponse)
	if err != nil {
		log.Fatalf("Error parsing response body: %v", err)
	}

	return &meResponse, nil
}

func (c *Client) ChangePassword(password, newPassword string) error {
	payload := map[string]string{
		"password":     password,
		"password_new": newPassword,
	}

	resp, err := c.Post("/api/auth/me/change_password/", payload)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("password change failed: check your current password or the new password")
	}

	return nil
}

func (c *Client) UpdateUsername(newUsername string) error {
	payload := map[string]string{
		"username": newUsername,
	}

	resp, err := c.Patch("/api/auth/me/", payload)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to update username")
	}

	return nil
}
