package gps

import (
	"time"

	"github.com/golang/geo/s2"
)

// Nothing should stop you from using a real GPS :D
type GPS interface {
	CurrentPos() s2.LatLng
	ID() string
}

var nowFunc = time.Now

// A GPS simulator
type SimGPS struct {
	id         string
	lw         LineWalker
	vel        float64
	lastReport time.Time
}

func NewSimGPS(id string, vel float64, lw LineWalker) GPS {
	return &SimGPS{
		id:         id,
		lw:         lw,
		vel:        vel,
		lastReport: nowFunc(),
	}
}

func (gps *SimGPS) ID() string {
	return gps.id
}

func (sim *SimGPS) CurrentPos() s2.LatLng {
	now := nowFunc()
	dist := (now.Sub(sim.lastReport).Hours()) * sim.vel

	sim.lastReport = now

	pt, _ := sim.lw.Walk(dist)
	return s2.LatLngFromPoint(pt)
}
