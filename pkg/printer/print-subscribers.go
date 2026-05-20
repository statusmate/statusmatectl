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
	verified := "no"
	if sub.Verified {
		verified = "yes"
	}

	_, err := fmt.Fprintf(w,
		"uuid=%s\nemail=%s\nverified=%s\nsubscribe_by_email=%v\nsubscribe_by_webhook=%v\ncreated_at=%s\n",
		uuid,
		sub.Email,
		verified,
		sub.SubscribeByEmail,
		sub.SubscribeByWebhook,
		formatTime(sub.CreatedAt),
	)
	return err
}
