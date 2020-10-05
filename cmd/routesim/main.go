package main

import (
	"encoding/json"
	"os"

	"github.com/gpontesss/routesim/cmd/routesim/internal/config"
)

const filePath = "samples/shpfile.json"

func main() {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	var cfg config.Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		panic(err)
	}

	sim, err := cfg.BuildRouteSim()
	if err != nil {
		panic(err)
	}

	if err := sim.Run(); err != nil {
		panic(err)
	}
}
