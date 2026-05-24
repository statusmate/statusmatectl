package printer

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/statusmate/statusmatectl/pkg/api"
)

func PrintTeams(w io.Writer, paginated *api.Paginated[api.Team], config *PrintTableConfig) error {
	if config.Format == PrintTableFormatJSON {
		return PrintAsJSON(w, paginated, config)
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	if config.PrintBlockTotal {
		fmt.Fprintf(tw, "total %d\n\n", paginated.Count)
	}

	if len(paginated.Results) == 0 {
		fmt.Fprintln(w, "No results to display.")
		return nil
	}

	fmt.Fprintln(tw, "ID\tName\tSlug\tRole\tCreated\t")
	for _, t := range paginated.Results {
		role := ""
		if t.TeamUser != nil {
			role = t.TeamUser.Role
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t\n",
			t.ID, t.Name, t.Slug, role, formatTime(t.CreatedAt))
	}

	return tw.Flush()
}

func PrintTeamInvites(w io.Writer, paginated *api.Paginated[api.TeamInvite], config *PrintTableConfig) error {
	if config.Format == PrintTableFormatJSON {
		return PrintAsJSON(w, paginated, config)
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	if config.PrintBlockTotal {
		fmt.Fprintf(tw, "total %d\n\n", paginated.Count)
	}

	if len(paginated.Results) == 0 {
		fmt.Fprintln(w, "No results to display.")
		return nil
	}

	fmt.Fprintln(tw, "Code\tEmail\tRole\tAccepted\tCreated\t")
	for _, inv := range paginated.Results {
		accepted := "-"
		if inv.AcceptedAt != nil {
			accepted = inv.AcceptedAt.Format("02 Jan 15:04")
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t\n",
			inv.Code, inv.Email, inv.Role, accepted, formatTime(inv.CreatedAt))
	}

	return tw.Flush()
}

func PrintTeamUsers(w io.Writer, paginated *api.Paginated[api.TeamUser], config *PrintTableConfig) error {
	if config.Format == PrintTableFormatJSON {
		return PrintAsJSON(w, paginated, config)
	}
	return PrintAsTable(w, paginated, config)
}

func PrintSummaryTeamInvite(w io.Writer, inv *api.TeamInvite) error {
	accepted := "-"
	if inv.AcceptedAt != nil {
		accepted = inv.AcceptedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	_, err := fmt.Fprintf(w,
		"uuid=%s\ncode=%s\nemail=%s\nrole=%s\nteam=%d\naccepted_at=%s\ncreated_at=%s\n",
		inv.UUID,
		inv.Code,
		inv.Email,
		inv.Role,
		inv.Team,
		accepted,
		formatTime(inv.CreatedAt),
	)
	return err
}
