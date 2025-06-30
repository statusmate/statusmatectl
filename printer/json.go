package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"statusmatectl/api"
)

func PrintAsJSON[T any](writer io.Writer, paginated *api.Paginated[T], config *PrintTableConfig) error {
	prettyJSON, err := json.MarshalIndent(paginated, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, err = writer.Write(prettyJSON)
	return err
}
