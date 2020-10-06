package gps

import (
	"time"

	"github.com/golang/geo/s2"
	geojson "github.com/paulmach/go.geojson"
)

// FreqEmitter makes a channel available that receives latlng positions formatted
// as GeoJSON points with any additional properties desired
type FreqEmitter struct {
	id         string
	curPosFunc func() s2.LatLng
	freq       time.Duration
	addPros    map[string]interface{}
	posChan    chan geojson.Feature
}

// NewFreqEmitter assembles a GPS position emitter
func NewFreqEmitter(gps GPS, freq time.Duration, props map[string]interface{}) *FreqEmitter {
	emitter := &FreqEmitter{
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
func (e *FreqEmitter) init() {
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
func (e *FreqEmitter) Positions() <-chan geojson.Feature {
	return e.posChan
}
