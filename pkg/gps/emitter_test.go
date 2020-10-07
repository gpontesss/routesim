package gps

import (
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/stretchr/testify/assert"
)

type fakeGPS struct {
	id       string
	curi     int
	posl     []s2.LatLng
	metadata map[string]interface{}
}

func FakeGPS(id string, posl ...s2.LatLng) GPS {
	return &fakeGPS{
		id:       id,
		curi:     0,
		posl:     posl,
		metadata: map[string]interface{}{},
	}
}

func (gps *fakeGPS) CurrentPos() Position {
	curll := gps.posl[gps.curi]
	gps.curi++
	return Position{
		GPS:    gps,
		At:     nowFunc(),
		LatLng: curll,
	}
}
func (gps *fakeGPS) ID() string                       { return gps.id }
func (gps *fakeGPS) Metadata() map[string]interface{} { return gps.metadata }

// Returns a function that returns a ticker that ticks n times
func TickerFunc(n int) func(time.Duration) *time.Ticker {
	tickc := make(chan time.Time)
	go func() {
		for i := 0; i < n; i++ {
			tickc <- time.Time{}
		}
		close(tickc)
	}()
	return func(_ time.Duration) *time.Ticker {
		return &time.Ticker{
			C: tickc,
		}
	}
}

func TestFreqEmitter(t *testing.T) {
	posl := []s2.LatLng{
		s2.LatLngFromDegrees(0, 0),
		s2.LatLngFromDegrees(0, 90),
		s2.LatLngFromDegrees(0, 180),
		s2.LatLngFromDegrees(0, -90),
		s2.LatLngFromDegrees(0, 0),
	}
	gps := FakeGPS("TEST1234", posl...)

	tickerFunc = TickerFunc(len(posl))
	emt := NewFreqEmitter(gps, time.Second)
	for _, pos := range posl {
		assert.True(t, pos.ApproxEqual((<-emt.Positions()).LatLng))
	}
}
