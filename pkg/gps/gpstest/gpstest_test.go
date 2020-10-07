package gpstest

import (
	"testing"

	"github.com/golang/geo/s2"
	"github.com/stretchr/testify/assert"
)

// Getting so meta...
func TestTestGPS(t *testing.T) {
	posl := []s2.LatLng{
		s2.LatLngFromDegrees(0, 0),
		s2.LatLngFromDegrees(90, 0),
		s2.LatLngFromDegrees(0, 180),
		s2.LatLngFromDegrees(-90, 0),
		s2.LatLngFromDegrees(0, 0),
	}
	gps := TestGPS("TEST1234", posl...)

	assert.Equal(t, "TEST1234", gps.ID())
	for _, pos := range posl {
		assert.True(t, pos.ApproxEqual(gps.CurrentPos().LatLng))
	}
}
