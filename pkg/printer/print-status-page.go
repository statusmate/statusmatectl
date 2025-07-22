package printer

import (
	"io"

	"github.com/statusmate/statusmatectl/pkg/api"
)

func PrintStatusPages(w io.Writer, paginated *api.Paginated[api.StatusPage], config *PrintTableConfig) error {
	if config.Format == PrintTableFormatTable {
		return PrintAsTable(w, paginated, config)
	}
	if config.Format == PrintTableFormatJSON {
		return PrintAsJSON(w, paginated, config)
	}

	return nil
}
