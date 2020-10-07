package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gpontesss/routesim/cmd/routesim/internal/config"
	"github.com/gpontesss/routesim/cmd/routesim/internal/routesim"
	"github.com/spf13/cobra"
)

// RouteSimCmd returns a routesim command
func RouteSimCmd() *cobra.Command {
	return routeSimCmd
}

var (
	cfgPath     string
	routeSimCmd = &cobra.Command{
		Use:   "routesim",
		Short: "GPS route simulator",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				cfg config.Config
				sim *routesim.RouteSim
			)
			cfgFile, err := os.Open(cfgPath)
			if err != nil {
				return fmt.Errorf("Error reading config file: %v", err)
			}
			if err := json.NewDecoder(cfgFile).Decode(&cfg); err != nil {
				return fmt.Errorf("Error loading config file: %v", err)
			}
			if sim, err = cfg.BuildRouteSim(); err != nil {
				return err
			}
			return sim.Run()
		},
	}
)

func init() {
	routeSimCmd.Flags().StringVarP(
		&cfgPath,
		"config",
		"c",
		"routesim.json",
		"Path to configuration file")
}
