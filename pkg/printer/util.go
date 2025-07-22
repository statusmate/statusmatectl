package printer

import (
	"bytes"
	"errors"
	"golang.org/x/term"
	"os"
	"time"
)

func formatTime(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format("02 Jan 15:04")
}

func nullOrValue(v *string) string {
	if v != nil {
		return *v
	}
	return "unknown"
}

func checkInteractiveTerminal() error {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return err
	}

	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return errors.New("this is not an interactive shell")
	}

	return nil
}

func termHeight(file *os.File) (int, error) {
	_, h, err := term.GetSize(int(file.Fd()))
	if err != nil {
		return -1, err
	}

	return h, nil
}

func lineCount(b []byte) int {
	return bytes.Count(b, []byte{'\n'})
}
