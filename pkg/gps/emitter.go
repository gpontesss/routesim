package gps

import (
	"time"
)

// FreqEmitter makes a channel available that receives latlng positions
type FreqEmitter struct {
	curPosFunc func() Position
	freq       time.Duration
	posChan    chan Position
}

// NewFreqEmitter assembles a GPS position emitter
func NewFreqEmitter(gps GPS, freq time.Duration) *FreqEmitter {
	emt := &FreqEmitter{
		curPosFunc: gps.CurrentPos,
		freq:       freq,
		// Should it be buffered?
		posChan: make(chan Position),
	}
	emt.init()
	return emt
}

// Initializes a goroutine for querying GPS's position with desired frequency
func (emt *FreqEmitter) init() {
	ticker := time.NewTicker(emt.freq)
	go func() {
		for range ticker.C {
			emt.posChan <- emt.curPosFunc()
		}
	}()
}

// Positions returns a channel that receives positions with desired frequency
func (emt *FreqEmitter) Positions() <-chan Position {
	return emt.posChan
}
