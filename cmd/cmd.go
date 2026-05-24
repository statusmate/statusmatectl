package cmd

import (
	"context"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

const Version = "0.0.1"

func init() {
	godotenv.Load()
}

func Run(args []string) error {
	RootCmd.SetArgs(args)

	defer func() {
		if x := recover(); x != nil {
			printPanic(x)

			os.Exit(1)
		}
	}()

	ctx := context.Background()

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
