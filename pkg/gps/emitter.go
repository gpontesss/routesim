package gps

import (
	"time"
)

// FreqEmitter makes a channel available that receives latlng positions
type FreqEmitter struct {
	curPosFunc func() Position
	posChan    chan Position
}

// NewFreqEmitter assembles a GPS position emitter
func NewFreqEmitter(gps GPS, freq time.Duration) *FreqEmitter {
	emt := &FreqEmitter{
		curPosFunc: gps.CurrentPos,
		// Should it be buffered?
		posChan: make(chan Position),
	}
	emt.init(freq)
	return emt
}

// Creates a ticker with a duration
var tickerFunc func(time.Duration) *time.Ticker

// Initializes a goroutine for querying GPS's position with desired frequency
func (emt *FreqEmitter) init(freq time.Duration) {
	ticker := tickerFunc(freq)
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
