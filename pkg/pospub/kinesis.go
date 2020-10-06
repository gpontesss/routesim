package pospub

import (
	"errors"

	geojson "github.com/paulmach/go.geojson"
)

type kinesisPosPub struct {
	streamName string
	// TODO: add AWS credentials as optional
}

// KinesisPosPublisher docs here
func KinesisPosPublisher(streamName string) PosPublisher {
	return &kinesisPosPub{streamName: streamName}
}

// PublishPos docs here
// TODO: deal with buffering and batch writing
func (p *kinesisPosPub) PublishPos(pos geojson.Feature) error {
	return errors.New("Not implemented")
}
