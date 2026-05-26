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
	Key               string `json:"key"`
	Created           string `json:"created"`
	User              int    `json:"user"`
	TwoFactorToken    string `json:"token"`
	TwoFactorRequired bool   `json:"-"`
}

type TwoFactorRequest struct {
	Code  string `json:"code"`
	Token string `json:"token"`
}

func (c *Client) Login(email string, password string) (*User, *AuthResponse, error) {
	authReq := AuthRequest{
		Username: email,
		Password: password,
	}

	resp, err := c.Post("/api/auth/signin/", authReq)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var authResponse AuthResponse
		if err = parseResponseBody(resp, &authResponse); err != nil {
			log.Fatalf("Failed to parse response: %v", err)
		}
		c.SetAuthToken(authResponse.Key)
		user, err := c.GetMe()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to received user: %v", err)
		}
		return user, &authResponse, nil

	case http.StatusCreated:
		var authResponse AuthResponse
		if err = parseResponseBody(resp, &authResponse); err != nil {
			log.Fatalf("Failed to parse 2FA response: %v", err)
		}
		authResponse.TwoFactorRequired = true
		return nil, &authResponse, nil

	default:
		return nil, nil, errors.New("failed to login, invalid credentials")
	}
}

func (c *Client) TwoFactorVerify(code, token string) (*User, *AuthResponse, error) {
	req := TwoFactorRequest{
		Code:  code,
		Token: token,
	}

	resp, err := c.Post("/api/auth/two_factor_verify/", req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, errors.New("invalid 2FA code")
	}

	var authResponse AuthResponse
	if err = parseResponseBody(resp, &authResponse); err != nil {
		log.Fatalf("Failed to parse 2FA verify response: %v", err)
	}

	c.SetAuthToken(authResponse.Key)

	user, err := c.GetMe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to received user: %v", err)
	}

	return user, &authResponse, nil
}
