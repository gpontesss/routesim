package routesim

import (
	"github.com/gpontesss/routesim/pkg/emitter"
	"github.com/gpontesss/routesim/pkg/pospub"
)

// RouteSim ingests and publishes simulated GPS position emmitions
type RouteSim struct {
	emitters  []*emitter.PosEmitter
	publisher pospub.PosPublisher
}

// NewRouteSim builds a RouteSim
func NewRouteSim(ems []*emitter.PosEmitter, pub pospub.PosPublisher) *RouteSim {
	return &RouteSim{
		emitters:  ems,
		publisher: pub,
	}
}

// Run starts RouteSim ingestion and publishing. It stops if any error occurs.
func (sim *RouteSim) Run() error {
	for {
		for _, em := range sim.emitters {
			select {
			case pos := <-em.Positions():
				if err := sim.publisher.PublishPos(pos); err != nil {
					return err
				}
			default:
				continue
			}
		}
	}
}
