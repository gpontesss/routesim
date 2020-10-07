package pospub

import (
	"github.com/gpontesss/routesim/pkg/gps"
)

// PosPublisher publishes a GPS position
type PosPublisher interface {
	PublishPos(gps.Position) error
}
