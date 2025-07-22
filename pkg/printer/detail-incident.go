package printer

import (
	"fmt"
	"io"
	"strings"

	"github.com/statusmate/statusmatectl/pkg/api"
)

func PrintSummaryIncident(w io.Writer, incident *api.Incident) error {
	var components []string

	createdAt := formatTime(incident.CreatedAt)
	uuid := nullOrValue(incident.UUID)
	name := incident.Title
	description := incident.Description
	componentList := strings.Join(components, ", ")
	status := string(incident.Status)

	summary := fmt.Sprintf(
		"uuid=%s\n"+
			"name=%s\n"+
			"description=%s\n"+
			"components=%s\n"+
			"status=%s\n"+
			"created_at=%s",
		uuid, name, description, componentList, status, createdAt,
	)

	_, err := w.Write([]byte(summary))
	return err
}
