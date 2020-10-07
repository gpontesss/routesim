package main

import (
	"os"

	"github.com/gpontesss/routesim/cmd/routesim/internal/cmd"
)

func main() {
	if err := cmd.RouteSimCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
