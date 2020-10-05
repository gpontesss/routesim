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
	// tells if it has crossed one of the line edges
	Walk(dist float64) (s2.Point, bool)
}

type backForthWalker struct {
	path    *s2.Polyline
	currPos float64
	len     s1.Angle
}

func BackForthWalker(path *s2.Polyline) LineWalker {
	return &BackForthLineWalker{
		path:    path,
		currPos: 0,
		len:     path.Length(),
	}
}

func (w *backForthWalker) Reset() {
	w.currPos = 0
}

func (w *backForthWalker) Walk(dist float64) (s2.Point, bool) {
	var pt s2.Point

	distFrac := float64(metersToS1Angle(dist) / w.len)
	crossedEdge := w.currPos < 1 && 1 <= w.currPos+distFrac

	if w.currPos += distFrac; 2 <= w.currPos {
		w.currPos -= 2
		crossedEdge = true
	}

	if w.currPos >= 1 {
		pt, _ = w.path.Interpolate(2 - w.currPos)
	} else {
		pt, _ = w.path.Interpolate(w.currPos)
	}

	return pt, crossedEdge
}

type restartWalker struct {
	path    *s2.Polyline
	currPos float64
	len     s1.Angle
}

func RestartWalker(path *s2.Polyline) LineWalker {
	return &restartWalker{
		path:    path,
		currPos: 0,
		len:     path.Length(),
	}
}

func (w *restartWalker) Reset() {
	w.currPos = 0
}

func (w *restartWalker) Walk(dist float64) (s2.Point, bool) {
	crossedEdge := false
	frac := float64(metersToS1Angle(dist) / w.len)

	if w.currPos+frac > 1 {
		crossedEdge = true
		w.currPos -= 1
	}

	pt, _ := w.path.Interpolate(frac)
	return pt, crossedEdge
}

func metersToS1Angle(m float64) s1.Angle {
	panic("Not implemented")
}
