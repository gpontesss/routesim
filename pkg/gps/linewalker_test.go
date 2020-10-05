package gps

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/golang/geo/s2"
	"github.com/stretchr/testify/assert"
)

/*
	How to describe it visually?
	It should represent the earth curvature
			 (90,0)
		   ____(1)____
		  /     |     \
	  (45,-90)  |   (45,90)
		|   \   |   /   |
	   (0)   \  |  /   (2)
	 (0,-90)  \ | /   (0,90)
*/
var mezzalunaPath = s2.PolylineFromLatLngs(
	[]s2.LatLng{
		s2.LatLngFromDegrees(0, -90),
		s2.LatLngFromDegrees(90, 0),
		s2.LatLngFromDegrees(0, 90),
	},
)

func TestWalkers(t *testing.T) {
	cases := map[LineWalker][]struct {
		dist        Distance
		endLatLng   s2.LatLng
		crossedEdge bool
	}{
		RestartWalker(mezzalunaPath): {
			{DistanceFromMeters(earthRadius * (math.Pi / 4)), s2.LatLngFromDegrees(45, -90), false},
			{DistanceFromMeters(earthRadius * (math.Pi / 4)), s2.LatLngFromDegrees(90, 0), false},
			{DistanceFromMeters(earthRadius * math.Pi), s2.LatLngFromDegrees(90, 0), true},
		},
		BackForthWalker(mezzalunaPath): {
			{DistanceFromMeters(earthRadius * (math.Pi / 2)), s2.LatLngFromDegrees(90, 0), false},
			{DistanceFromMeters(earthRadius * (math.Pi / 4)), s2.LatLngFromDegrees(45, 90), false},
			{DistanceFromMeters(earthRadius * (math.Pi / 2)), s2.LatLngFromDegrees(45, 90), true},
		},
	}

	for lw, walks := range cases {
		for _, walk := range walks {
			t.Run(
				fmt.Sprintf("%s/Distance-%.6f/Stop-%v", reflect.TypeOf(lw), walk.dist, walk.endLatLng),
				func(t *testing.T) {
					ll, ce := lw.Walk(walk.dist)
					assert.True(t, walk.endLatLng.ApproxEqual(ll), "Expected: %v/Result: %v", walk.endLatLng, ll)
					assert.Equal(t, walk.crossedEdge, ce)
				},
			)
		}
	}
}

func TestMetersToS1Angle(t *testing.T) {
	cases := []struct {
		in     float64
		result Distance
	}{
		{earthRadius, 1},
		{earthRadius * math.Pi, math.Pi},
	}

	for _, c := range cases {
		assert.Equal(t, c.result, DistanceFromMeters(c.in))
	}
}
