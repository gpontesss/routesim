package emitter

import (
	"time"

	"github.com/golang/geo/s2"
	"github.com/gpontesss/routesim/pkg/gps"
	geojson "github.com/paulmach/go.geojson"
)

// Emitter makes a channel available that receives latlng positions formatted
// as GeoJSON points with any additional properties desired
type PosEmitter struct {
	id         string
	curPosFunc func() s2.LatLng
	freq       time.Duration
	addPros    map[string]interface{}
	posChan    chan geojson.Feature
}

// GPSEmitter assembles a GPS position emitter
func GPSPosEmitter(gps gps.GPS, freq time.Duration, props map[string]interface{}) *PosEmitter {
	emitter := &PosEmitter{
		id:         gps.ID(),
		curPosFunc: gps.CurrentPos,
		freq:       freq,
		// Should it be buffered?
		posChan: make(chan geojson.Feature),
	}
	emitter.init()
	return emitter
}

// Initializes a goroutine for querying GPS's position with desired frequency
func (e *PosEmitter) init() {
	ticker := time.NewTicker(e.freq)
	go func() {
		for range ticker.C {
			pos := e.curPosFunc()
			e.posChan <- geojson.Feature{
				ID:         e.id,
				Type:       "Feature",
				Properties: e.addPros,
				Geometry: &geojson.Geometry{
					Type: geojson.GeometryPoint,
					Point: []float64{
						pos.Lat.Degrees(),
						pos.Lng.Degrees(),
					},
				},
			}
		}
	}()
}

// Positions returns a channel that receives positions with desired frequency
func (e *PosEmitter) Positions() <-chan geojson.Feature {
	return e.posChan
}
