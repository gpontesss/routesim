package routesim

import (
	"time"

	"github.com/gpontesss/routesim/pkg/gps"
)

// FreqEmitter makes a channel available that receives latlng positions
type FreqEmitter struct {
	curPosFunc func() gps.Position
	posChan    chan gps.Position
}

// NewFreqEmitter assembles a GPS position emitter
func NewFreqEmitter(gps gps.GPS, freq time.Duration) *FreqEmitter {
	return FreqEmitterWithTicker(gps, tickerFunc(freq))
}

// FreqEmitterWithTicker docs here
func FreqEmitterWithTicker(gpz gps.GPS, ticker *time.Ticker) *FreqEmitter {
	emt := &FreqEmitter{
		curPosFunc: gpz.CurrentPos,
		// Should it be buffered?
		posChan: make(chan gps.Position),
	}
	emt.init(ticker)
	return emt
}

// Creates a ticker with a duration
var tickerFunc func(time.Duration) *time.Ticker

// Initializes a goroutine for querying GPS's position with desired frequency
func (emt *FreqEmitter) init(ticker *time.Ticker) {
	go func() {
		for range ticker.C {
			emt.posChan <- emt.curPosFunc()
		}
	}()
}

// Positions returns a channel that receives positions with desired frequency
func (emt *FreqEmitter) Positions() <-chan gps.Position {
	return emt.posChan
}
