package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/list"
	"golang.org/x/term"
)

type LogEntry struct {
	At       time.Time `json:"at"`
	UUID     string    `json:"uuid"`
	Object   string    `json:"object"` // incident | maintenance
	Title    string    `json:"title"`
	Status   string    `json:"status"`
	Desc     string    `json:"description,omitempty"`
	ParentID int       `json:"parent_id,omitempty"`
}

func PrintLogs(w io.Writer, entries []LogEntry, config *PrintTableConfig) error {
	if config.Format == PrintTableFormatJSON {
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, string(data))
		return err
	}

	if config.Format == PrintTableFormatTimeline {
		return printLogsTimeline(w, entries)
	}

	if len(entries) == 0 {
		fmt.Fprintln(w, "No log entries.")
		return nil
	}

	color := IsTerminal(w)

	for i, e := range entries {
		if i > 0 {
			fmt.Fprintln(w)
		}

		shortID := e.UUID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}

		header := "update " + shortID
		if color {
			fmt.Fprintf(w, "\033[33m%s\033[0m\n", header)
		} else {
			fmt.Fprintln(w, header)
		}

		fmt.Fprintf(w, "Date:   %s\n", e.At.Local().Format("Mon, 02 Jan 2006 15:04"))
		fmt.Fprintln(w)

		fmt.Fprintf(w, "    Object: %s\n", e.Object)
		fmt.Fprintf(w, "    Status: %s\n", logShortStatus(e.Status))
		fmt.Fprintf(w, "    Title: %s\n", e.Title)
		if e.Desc != "" {
			desc := strings.ReplaceAll(strings.TrimSpace(e.Desc), "\n", " ")
			fmt.Fprintf(w, "    Message: %s\n", desc)
		}
	}

	return nil
}

func printLogsTimeline(w io.Writer, entries []LogEntry) error {
	if len(entries) == 0 {
		fmt.Fprintln(w, "No log entries.")
		return nil
	}

	color := IsTerminal(w)

	type group struct {
		object  string
		title   string
		entries []LogEntry
	}

	var order []int
	groups := map[int]*group{}

	for _, e := range entries {
		if _, ok := groups[e.ParentID]; !ok {
			order = append(order, e.ParentID)
			groups[e.ParentID] = &group{object: e.Object, title: e.Title}
		}
		groups[e.ParentID].entries = append(groups[e.ParentID].entries, e)
	}

	for _, g := range groups {
		sort.Slice(g.entries, func(i, j int) bool {
			return g.entries[i].At.Before(g.entries[j].At)
		})
	}

	lw := list.NewWriter()
	lw.SetStyle(list.StyleConnectedRounded)

	for _, id := range order {
		g := groups[id]

		header := fmt.Sprintf("[%s] %s", g.object, g.title)
		if color {
			header = timelineColorObject(g.object) + header + "\033[0m"
		}
		lw.AppendItem(header)
		lw.Indent()

		for _, e := range g.entries {
			ts := e.At.Local().Format("Jan 02 15:04")
			status := logShortStatus(e.Status)
			line := fmt.Sprintf("%-14s  %-14s", ts, status)
			if e.Desc != "" {
				desc := strings.ReplaceAll(strings.TrimSpace(e.Desc), "\n", " ")
				if len([]rune(desc)) > 60 {
					desc = string([]rune(desc)[:57]) + "..."
				}
				line += "  " + desc
			}
			lw.AppendItem(line)
		}

		lw.UnIndent()
	}

	fmt.Fprintln(w, lw.Render())
	return nil
}

func timelineColorObject(object string) string {
	switch object {
	case "incident":
		return "\033[31m" // red
	case "maintenance":
		return "\033[34m" // blue
	default:
		return "\033[33m" // yellow
	}
}

func logShortStatus(s string) string {
	s = strings.TrimPrefix(s, "incident_")
	s = strings.TrimPrefix(s, "maintenance_")
	return s
}

func IsTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}
