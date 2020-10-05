package gps

import (
	"time"

	"github.com/golang/geo/s2"
)

// GPS describes a GPS behavior
// Nothing should stop you from using a real GPS :D
type GPS interface {
	// CurrentPos should return GPS's current position
	CurrentPos() s2.LatLng
	// ID should return GPS's unique identifier
	ID() string
}

var nowFunc = time.Now

// SimGPS simulates a real GPS that walks a line
type SimGPS struct {
	id         string
	lw         LineWalker
	vel        float64
	lastReport time.Time
}

// NewSimGPS creates a GPS simulator that walks a line with a constant velocity.
// Velocity is given by m/s.
func NewSimGPS(id string, vel float64, lw LineWalker) GPS {
	return &SimGPS{
		id:         id,
		lw:         lw,
		vel:        vel,
		lastReport: nowFunc(),
	}
}

// ID returns the GPS ID
func (gps *SimGPS) ID() string {
	return gps.id
}

// CurrentPos returns the GPS current position
func (gps *SimGPS) CurrentPos() s2.LatLng {
	now := nowFunc()
	dist := now.Sub(gps.lastReport).Seconds() * gps.vel

	gps.lastReport = now

	ll, _ := gps.lw.Walk(DistanceFromMeters(dist))
	return ll
}
