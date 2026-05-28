package editor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const DefaultEditor = "vim"

type PreferredEditorResolver func() string

func GetPreferredEditorFromEnvironment() string {
	editor := os.Getenv("EDITOR")

	if editor == "" {
		return DefaultEditor
	}

	return editor
}

func resolveEditorArguments(executable string, filename string) []string {
	args := []string{filename}

	base := strings.ToLower(filepath.Base(executable))
	if strings.Contains(executable, "Visual Studio Code.app") || base == "code" || base == "zed" {
		args = append([]string{"--wait"}, args...)
	}

	return args
}

func OpenFileInEditor(filename string, resolveEditor PreferredEditorResolver) error {
	executable, err := exec.LookPath(resolveEditor())
	if err != nil {
		return err
	}

	args := resolveEditorArguments(executable, filename)

	cmd := exec.Command(executable, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func CaptureInputFromEditor(data []byte) ([]byte, error) {
	file, err := os.CreateTemp(os.TempDir(), "*")
	if err != nil {
		return []byte{}, err
	}

	filename := file.Name()

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return nil, fmt.Errorf("failed to save data: %v", err)
	}

	defer os.Remove(filename)

	if err = file.Close(); err != nil {
		return nil, err
	}

	if err = OpenFileInEditor(filename, GetPreferredEditorFromEnvironment); err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
