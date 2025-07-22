package cmd

import (
	"context"
	"net/url"
	"os"
)

const Version = "0.0.1"

func Run(args []string) error {
	RootCmd.SetArgs(args)

	defer func() {
		if x := recover(); x != nil {
			printPanic(x)

			os.Exit(1)
		}
	}()

	ctx := context.Background()

	ctx = context.WithValue(ctx, "authrc", "")
	ctx = context.WithValue(ctx, "UseStatusPage", "")
	ctx = context.WithValue(ctx, "UseReleasePage", "")

	ctx.Value("Token")
	ctx.Value("UseStatusPage")
	ctx.Value("UseReleasePage")

	return RootCmd.ExecuteContext(ctx)
}

func printPanic(_ any) {
	u, err := url.Parse("https://github.com/statusmate/statusmate/issues/new")
	if err != nil {
		panic(err)
	}

	q := u.Query()
	q.Add("title", "[statusmate] [bug] panic on v"+Version)
	q.Add("body", "<!--- Please provide the stack trace -->\n")
	u.RawQuery = q.Encode()
}
