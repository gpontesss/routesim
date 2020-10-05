package publisher

import (
	"errors"

	geojson "github.com/paulmach/go.geojson"
)

type kinesisPublisher struct {
	streamName string
}

func NewKinesisPublisher(streamName string) Publisher {
	return &kinesisDriver{streamName: streamName}
}

// TODO: deal with buffering and batch writing
func (drv *kinesisDriver) PublishPosition(positions geojson.Feature) error {
	return errors.New("Not implemented")
}
