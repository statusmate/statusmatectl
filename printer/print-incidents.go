package printer

import (
	"io"
	"statusmatectl/api"
)

func PrintIncidents(w io.Writer, paginated *api.Paginated[api.Incident], config *PrintTableConfig) error {
	if config.Format == PrintTableFormatTable {
		return PrintAsTable(w, paginated, config)
	}

	if config.Format == PrintTableFormatJSON {
		return PrintAsJSON(w, paginated, config)
	}

	return nil
}
