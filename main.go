package main

import (
	"os"
	"statusmatectl/cmd"
)

func main() {
	if err := cmd.Run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
