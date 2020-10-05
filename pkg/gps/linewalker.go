package gps

import (
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

// LineWalker walks a line
type LineWalker interface {
	// Resets current position to the start of the line
	Reset()
	// Walks a distance across a line and returns the point it stopped, and
	// tells if it has crossed one of the line edges. The distance is given in
	// meters.
	Walk(dist Distance) (s2.LatLng, bool)
}

// Distance docs here
type Distance s1.Angle

// In meters
const earthRadius = 6_371_000.0

// DistanceFromMeters docs here
func DistanceFromMeters(m float64) Distance {
	return Distance(m / earthRadius)
}

type backForthWalker struct {
	path    *s2.Polyline
	currPos float64
	len     Distance
}

// BackForthWalker creates a LineWalker that traverses a line forward and, when
// it reaches the end, goes back in reverse. The path should have lat lng
// coordinates.
func BackForthWalker(path *s2.Polyline) LineWalker {
	return &backForthWalker{
		path:    path,
		currPos: 0,
		len:     Distance(path.Length()),
	}
}

// Reset resets the LineWalker position to the start of the line
func (w *backForthWalker) Reset() {
	w.currPos = 0
}

// Walk walks a distance along the line. The distance should be given in meters.
func (w *backForthWalker) Walk(dist Distance) (s2.LatLng, bool) {
	distFrac := float64(dist / w.len)
	crossedEdge := w.currPos < 1 && 1 <= w.currPos+distFrac

	if w.currPos += distFrac; 2 <= w.currPos {
		w.currPos -= 2.0
		crossedEdge = true
	}

	var pt s2.Point
	if w.currPos >= 1 {
		pt, _ = w.path.Interpolate(2 - w.currPos)
	} else {
		pt, _ = w.path.Interpolate(w.currPos)
	}

	return s2.LatLngFromPoint(pt), crossedEdge
}

type restartWalker struct {
	path    *s2.Polyline
	currPos float64
	len     Distance
}

// RestartWalker creates a LineWalker that traverses a line forward and, when
// it reaches the end, goes back to the start. The path should have lat lng
// coordinates.
func RestartWalker(path *s2.Polyline) LineWalker {
	return &restartWalker{
		path:    path,
		currPos: 0,
		len:     Distance(path.Length()),
	}
}

// Reset resets the LineWalker position to the start of the line
func (w *restartWalker) Reset() {
	w.currPos = 0
}

// Walk walks a distance along the line. The distance should be given in meters.
func (w *restartWalker) Walk(dist Distance) (s2.LatLng, bool) {
	crossedEdge := false
	distFrac := float64(dist / w.len)

	w.currPos += distFrac
	if w.currPos > 1 {
		crossedEdge = true
		w.currPos -= 1.0
	}

	pt, _ := w.path.Interpolate(w.currPos)
	return s2.LatLngFromPoint(pt), crossedEdge
}
