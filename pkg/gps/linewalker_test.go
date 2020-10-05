package gps

import (
	"fmt"
	"math"
	"testing"

	"github.com/golang/geo/s2"
	"github.com/stretchr/testify/assert"
)

func TestRestartWalker(t *testing.T) {
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
	path := s2.PolylineFromLatLngs(
		[]s2.LatLng{
			s2.LatLngFromDegrees(0, -90),
			s2.LatLngFromDegrees(90, 0),
			s2.LatLngFromDegrees(0, -90),
		},
	)

	lw := RestartWalker(path)

	cases := []struct {
		dist        Distance
		ll          s2.LatLng
		crossedEdge bool
	}{
		{DistanceFromMeters(earthRadius * (math.Pi / 4)), s2.LatLngFromDegrees(45, -90), false},
		{DistanceFromMeters(earthRadius * (math.Pi / 4)), s2.LatLngFromDegrees(90, 0), false},
		{DistanceFromMeters(4 * (earthRadius * (math.Pi / 4))), s2.LatLngFromDegrees(90, 0), true},
	}

	for _, c := range cases {
		t.Run(fmt.Sprint(c.dist, c.ll), func(t *testing.T) {
			ll, ce := lw.Walk(c.dist)
			assert.True(t, c.ll.ApproxEqual(ll), "Expected: %v/Result: %v", c.ll, ll)
			assert.Equal(t, c.crossedEdge, ce)
		})
	}
}

func TestBackAndForthWalker(t *testing.T) {
	panic("TODO")
}

func TestMetersToS1Angle(t *testing.T) {
	cases := []struct {
		in     float64
		result Distance
	}{
		{earthRadius * math.Pi, math.Pi},
	}

	for _, c := range cases {
		assert.Equal(t, c.result, DistanceFromMeters(c.in))
	}
}
