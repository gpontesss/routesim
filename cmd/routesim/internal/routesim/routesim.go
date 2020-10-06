package routesim

import (
	"github.com/gpontesss/routesim/pkg/gps"
	"github.com/gpontesss/routesim/pkg/pospub"
)

// RouteSim ingests and publishes simulated GPS position emmitions
type RouteSim struct {
	emitters  []*gps.FreqEmitter
	publisher pospub.PosPublisher
}

// NewRouteSim builds a RouteSim
func NewRouteSim(ems []*gps.FreqEmitter, pub pospub.PosPublisher) *RouteSim {
	return &RouteSim{
		emitters:  ems,
		publisher: pub,
	}
}

// Run starts RouteSim ingestion and publishing. It stops if any error occurs.
func (sim *RouteSim) Run() error {
	for {
		for _, emt := range sim.emitters {
			select {
			case pos := <-emt.Positions():
				if err := sim.publisher.PublishPos(pos); err != nil {
					return err
				}
			default:
				continue
			}
		}
	}
}
