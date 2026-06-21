package main

import (
	"os"

	"github.com/Omochice/glab-dep/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
