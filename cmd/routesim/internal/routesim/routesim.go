package routesim

import (
	"github.com/gpontesss/routesim/pkg/emitter"
	"github.com/gpontesss/routesim/pkg/publisher"
)

type RouteSim struct {
	emitters  []*emitter.PosEmitter
	publisher publisher.Publisher
}

func (sim *RouteSim) Run() error {
	for {
		for _, em := range sim.emitters {
			select {
			case pos := <-em.Positions():
				if err := sim.publisher(pos); err != nil {
					return err
				}
			default:
				continue
			}
		}
	}
}
