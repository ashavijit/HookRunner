package main

import (
	"os"

	"github.com/ashavijit/hookrunner/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
