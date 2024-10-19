package api

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type AuthRC struct {
	Username string
	Email    string
	Token    string
}

// sanitizeServer преобразует строку server в допустимый формат для файловой системы
func sanitizeServer(server string) string {
	// Удаляем протокол (http:// или https://)
	server = strings.TrimPrefix(server, "http://")
	server = strings.TrimPrefix(server, "https://")

	// Заменяем недопустимые символы
	replacer := strings.NewReplacer("/", "_", ":", "_")
	return replacer.Replace(server)
}

// SaveAuthRC сохраняет auth data в файл ~/.statusmate/{server}/authrc
func (c *Client) SaveAuthRC(auth *AuthResponse, user *User) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not find home directory: %w", err)
	}
	domain := sanitizeServer(c.BaseURL)
	authFilePath := filepath.Join(homeDir, ".statusmate", domain, "authrc")

	// Ensure the directory exists
	err = os.MkdirAll(filepath.Dir(authFilePath), 0700)
	if err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}

	// Write data
	file, err := os.Create(authFilePath)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()

	// Writing content to file
	content := fmt.Sprintf(
		"username=%s\nemail=%s\ntoken=%s\n",
		user.Username, user.Email, auth.Key,
	)
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}

	return nil
}

func (c *Client) LoadAuthRC(server string) (*AuthRC, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not find home directory: %w", err)
	}

	domain := sanitizeServer(c.BaseURL)
	authFilePath := filepath.Join(homeDir, ".statusmate", domain, "authrc")

	file, err := os.Open(authFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	authRC := &AuthRC{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]

		switch key {
		case "username":
			authRC.Username = value
		case "email":
			authRC.Email = value
		case "token":
			authRC.Token = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return authRC, nil
}
