package pospub

import (
	"errors"

	geojson "github.com/paulmach/go.geojson"
)

type kinesisPosPub struct {
	streamName string
	// TODO: add AWS credentials as optional
}

func KinesisPosPublisher(streamName string) PosPublisher {
	return &kinesisPosPub{streamName: streamName}
}

// TODO: deal with buffering and batch writing
func (p *kinesisPosPub) PublishPos(pos geojson.Feature) error {
	return errors.New("Not implemented")
}