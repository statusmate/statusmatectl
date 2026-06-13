package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	statusmateDir = ".st4"
)

type AuthRC struct {
	API                string         `json:"api"`
	Token              string         `json:"token"`
	DefaultStatusPage  string         `json:"default_status_page"`
	DefaultReleasePage string         `json:"default_release_page"`
	DefaultTeam        int            `json:"default_team,omitempty"`
	RecentPages        []string       `json:"recent_pages,omitempty"` // status-page slugs, oldest first, newest last, max maxRecentPages
}

const maxRecentPages = 5

// RecordPageVisit moves slug to the most-recent slot, keeping at most
// maxRecentPages entries (oldest first, newest last).
func (rc *AuthRC) RecordPageVisit(slug string) {
	if slug == "" {
		return
	}
	out := rc.RecentPages[:0]
	for _, s := range rc.RecentPages {
		if s != slug { // dedup: drop existing occurrence
			out = append(out, s)
		}
	}
	out = append(out, slug)
	if len(out) > maxRecentPages {
		out = out[len(out)-maxRecentPages:]
	}
	rc.RecentPages = out
}

func sanitizeDomain(server string) string {
	server = strings.TrimPrefix(server, "http://")
	server = strings.TrimPrefix(server, "https://")
	replacer := strings.NewReplacer("/", "_", ":", "_")
	return replacer.Replace(server)
}

func CheckDir(domain string, dest string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find home directory: %w", err)
	}
	filename := filepath.Join(homeDir, statusmateDir, sanitizeDomain(domain), dest)
	err = os.MkdirAll(filepath.Dir(filename), 0700)
	if err != nil {
		return "", fmt.Errorf("could not create directory: %w", err)
	}
	return filename, nil
}

func FromContext(ctx context.Context) (*AuthRC, bool) {
	v, ok := ctx.Value("authRc").(*AuthRC)
	return v, ok
}

func NewAuthRC(auth *AuthResponse) *AuthRC {
	return &AuthRC{
		Token: auth.Key,
	}
}

func SaveAuthRC(domain string, authRC *AuthRC) error {
	filename, err := CheckDir(domain, "authrc")
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
	filename, err := CheckDir(domain, "authrc")
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
