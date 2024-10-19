package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Key     string `json:"key"`
	Created string `json:"created"`
	User    int    `json:"user"`
}

func (c *Client) Login(email string, password string) (*User, error) {
	authReq := AuthRequest{
		Username: email,
		Password: password,
	}

	resp, err := c.Post("/api/auth/signin/", authReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to login, invalid credentials")
	}

	var authResponse AuthResponse
	err = parseResponseBody(resp, &authResponse)
	if err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	c.SetAuthToken(authResponse.Key)

	user, err := c.getMe()

	if err != nil {
		log.Fatalf("Failed to received user: %v", err)
	}

	err = c.SaveAuthRC(&authResponse, user)
	if err != nil {
		fmt.Errorf("failed save token to file")
	}

	return user, nil
}
