package main

import (
	"os"

	"github.com/statusmate/statusmatectl/cmd"
)

var version string

func main() {
	if version != "" {
		cmd.SetVersion(version)
	}
	if err := cmd.Run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
