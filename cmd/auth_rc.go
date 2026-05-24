package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/statusmate/statusmatectl/pkg/api"
)

func FromContext(ctx context.Context) (*api.AuthRC, bool) {
	v, ok := ctx.Value("authRc").(*api.AuthRC)
	return v, ok
}

func NewAuthRC(auth *api.AuthResponse) *api.AuthRC {
	return &api.AuthRC{
		Token: auth.Key,
	}
}

func SaveAuthRC(domain string, authRC *api.AuthRC) error {
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

func LoadAuthRC(domain string) (*api.AuthRC, error) {
	filename, err := checkDir(domain, "authrc")
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	authRC := &api.AuthRC{}

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
