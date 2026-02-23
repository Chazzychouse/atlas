package main

import (
	"os"

	"github.com/chazzychouse/atlas/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
