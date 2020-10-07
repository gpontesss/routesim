package routesim

import (
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/gpontesss/routesim/pkg/gps/gpstest"
	"github.com/stretchr/testify/assert"
)

// Ticker returns a ticker that ticks n times
func Ticker(n int) *time.Ticker {
	tickc := make(chan time.Time)
	go func() {
		for i := 0; i < n; i++ {
			tickc <- time.Time{}
		}
		close(tickc)
	}()
	return &time.Ticker{
		C: tickc,
	}
}

// TickerFunc returns a function that returns a ticker that ticks n times
func TickerFunc(n int) func(time.Duration) *time.Ticker {
	return func(_ time.Duration) *time.Ticker {
		return Ticker(n)
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
	gps := gpstest.TestGPS("TEST1234", posl...)

	tickerFunc = TickerFunc(len(posl))
	emt := NewFreqEmitter(gps, time.Second)
	for _, pos := range posl {
		assert.True(t, pos.ApproxEqual((<-emt.Positions()).LatLng))
	}
}
