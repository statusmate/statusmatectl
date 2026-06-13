package printer

import (
	"fmt"
	"io"

	"github.com/statusmate/statusmatectl/pkg/api"
)

func PrintSubscribers(w io.Writer, paginated *api.Paginated[api.Subscriber], config *PrintTableConfig) error {
	if config.Format == PrintTableFormatJSON {
		return PrintAsJSON(w, paginated, config)
	}

	return PrintAsTable(w, paginated, config)
}

func PrintSummarySubscriber(w io.Writer, sub *api.Subscriber) error {
	uuid := nullOrValue(sub.UUID)
	confirmed := "no"
	if sub.Confirmed {
		confirmed = "yes"
	}

	_, err := fmt.Fprintf(w,
		"uuid=%s\nemail=%s\nconfirmed=%s\nsubscribe_by_email=%v\nsubscribe_by_webhook=%v\ncreated_at=%s\n",
		uuid,
		sub.Email,
		confirmed,
		sub.SubscribeByEmail,
		sub.SubscribeByWebhook,
		formatTime(sub.CreatedAt),
	)
	return err
}
