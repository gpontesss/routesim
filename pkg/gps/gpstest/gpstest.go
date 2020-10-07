package gpstest

import (
	"time"

	"github.com/golang/geo/s2"
	"github.com/gpontesss/routesim/pkg/gps"
)

type testGPS struct {
	id       string
	curi     int
	posl     []s2.LatLng
	metadata map[string]interface{}
}

// TestGPS docs here
func TestGPS(id string, posl ...s2.LatLng) gps.GPS {
	return &testGPS{
		id:       id,
		curi:     0,
		posl:     posl,
		metadata: map[string]interface{}{},
	}
}

func (tgps *testGPS) CurrentPos() gps.Position {
	curll := tgps.posl[tgps.curi]
	tgps.curi++
	return gps.Position{
		GPS:    tgps,
		At:     time.Now(),
		LatLng: curll,
	}
}
func (tgps *testGPS) ID() string                       { return tgps.id }
func (tgps *testGPS) Metadata() map[string]interface{} { return tgps.metadata }
