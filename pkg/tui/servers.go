package tui

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/statusmate/statusmatectl/pkg/api"
)

const st4Dir = ".st4"

func listAvailableServers() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(filepath.Join(homeDir, st4Dir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var servers []string
	for _, e := range entries {
		if e.IsDir() {
			servers = append(servers, e.Name())
		}
	}
	return servers, nil
}

func loadServerAuthRC(domain string) (*api.AuthRC, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	filename := filepath.Join(homeDir, st4Dir, sanitizeDomain(domain), "authrc")
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var rc api.AuthRC
	if err := json.Unmarshal(data, &rc); err != nil {
		return nil, err
	}
	return &rc, nil
}

func loadServerClient(domain string, parent *api.Client) (*api.Client, error) {
	rc, err := loadServerAuthRC(domain)
	if err != nil {
		return nil, err
	}
	c := api.NewClient(domain, parent.Logger())
	c.AuthRC = rc
	return c, nil
}

func sanitizeDomain(domain string) string {
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	replacer := strings.NewReplacer("/", "_", ":", "_")
	return replacer.Replace(domain)
}
