package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/statusmate/statusmatectl/pkg/api"
)

type AuthRC struct {
	API                string `json:"api"`
	Token              string `json:"token"`
	DefaultStatusPage  string `json:"default_status_page"`
	DefaultReleasePage string `json:"default_release_page"`
}

func FromContext(ctx context.Context) (*AuthRC, bool) {
	v, ok := ctx.Value("authRc").(*AuthRC)
	return v, ok
}

func NewAuthRC(auth *api.AuthResponse) *AuthRC {
	return &AuthRC{
		Token: auth.Key,
	}
}

func SaveAuthRC(domain string, authRC *AuthRC) error {
	filename, err := checkDir(domain, "authrc")
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(authRC)
	if err != nil {
		return err
	}

	_, err = file.Write(data)

	return err
}

func LoadAuthRC(domain string) (*AuthRC, error) {
	filename, err := checkDir(domain, "authrc")
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	authRC := &AuthRC{}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	err = json.Unmarshal(data, authRC)
	if err != nil {
		return nil, err
	}

	return authRC, nil
}
