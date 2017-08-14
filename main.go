package main

import (
	"os"

	"github.com/kesslerm/syslogbeat/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
