package printer

import (
	"io"
	"statusmatectl/api"
)

func PrintMaintenances(w io.Writer, paginated *api.Paginated[api.Maintenance], config *PrintTableConfig) error {
	if config.Format == PrintTableFormatTable {
		return PrintAsTable(w, paginated, config)
	}

	if config.Format == PrintTableFormatJSON {
		return PrintAsJSON(w, paginated, config)
	}

	return nil
}
