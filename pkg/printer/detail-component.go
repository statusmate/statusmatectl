package printer

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/statusmate/statusmatectl/pkg/api"
)

func PrintDetailComponent(w io.Writer, comp *api.Component, format string) error {
	if format == PrintTableFormatJSON {
		data, err := json.MarshalIndent(comp, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON: %w", err)
		}
		_, err = fmt.Fprintln(w, string(data))
		return err
	}

	color := IsTerminal(w)

	shortID := ""
	if comp.UUID != nil {
		shortID = *comp.UUID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
	}

	header := "component " + shortID
	if color {
		fmt.Fprintf(w, "\033[36m%s\033[0m\n", header)
	} else {
		fmt.Fprintln(w, header)
	}

	if comp.UUID != nil {
		fmt.Fprintf(w, "UUID:        %s\n", *comp.UUID)
	}
	fmt.Fprintf(w, "Name:        %s\n", comp.Name)
	fmt.Fprintf(w, "Impact:      %s\n", comp.Impact)
	fmt.Fprintf(w, "Enabled:     %v\n", comp.Enabled)
	if comp.Description != "" {
		fmt.Fprintf(w, "Description: %s\n", comp.Description)
	}
	fmt.Fprintf(w, "Private:     %v\n", comp.Private)
	fmt.Fprintf(w, "Histogram:   %v\n", comp.Histogram)
	if comp.Uptime != "" {
		fmt.Fprintf(w, "Uptime:      %s\n", comp.Uptime)
	}
	if comp.CreatedAt != nil {
		fmt.Fprintf(w, "CreatedAt:   %s\n", comp.CreatedAt.Local().Format("Mon, 02 Jan 2006 15:04:05"))
	}
	if comp.UpdatedAt != nil {
		fmt.Fprintf(w, "UpdatedAt:   %s\n", comp.UpdatedAt.Local().Format("Mon, 02 Jan 2006 15:04:05"))
	}

	return nil
}

func PrintSummaryComponent(w io.Writer, comp *api.Component) {
	uuid := ""
	if comp.UUID != nil {
		uuid = *comp.UUID
	}
	fmt.Fprintf(w, "uuid=%s\nname=%s\nimpact=%s\nenabled=%v\n",
		uuid, comp.Name, comp.Impact, comp.Enabled)
}
