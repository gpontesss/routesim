package gps

import (
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
)

// GPS describes a GPS behavior
// Nothing should stop you from using a real GPS :D
type GPS interface {
	// ID should return GPS's unique identifier
	ID() string
	// CurrentPos should return GPS's current position
	CurrentPos() Position
	// Metadata describes any additional information about the device
	Metadata() map[string]interface{}
}

// Position gathers a lat lng position with its time of occurrence, and a
// reference to the GPS that generated it.
type Position struct {
	s2.LatLng
	At  time.Time
	GPS GPS
}

// Gets the current moment of time
var nowFunc = time.Now

// SimGPS simulates a real GPS that walks a line
type SimGPS struct {
	id         string
	lw         LineWalker
	vel        float64
	lastReport time.Time
	metadata   map[string]interface{}
}

// NewSimGPS creates a GPS simulator that walks a line with a constant velocity.
// Velocity is given by m/s.
func NewSimGPS(vel float64, lw LineWalker, metadata map[string]interface{}) GPS {
	return &SimGPS{
		id:         uuid.New().String(),
		lw:         lw,
		vel:        vel,
		lastReport: nowFunc(),
		metadata:   metadata,
	}
}

// ID returns the GPS' ID
func (gps *SimGPS) ID() string {
	return gps.id
}

// Metadata returns simulated device metadata
func (gps *SimGPS) Metadata() map[string]interface{} {
	return gps.metadata
}

// CurrentPos returns the GPS' current position
func (gps *SimGPS) CurrentPos() Position {
	now := nowFunc()
	dist := now.Sub(gps.lastReport).Seconds() * gps.vel

	gps.lastReport = now

	ll, _ := gps.lw.Walk(DistanceFromMeters(dist))
	return Position{LatLng: ll, GPS: gps, At: now}
}
