package data

import (
	"errors"
)

type kinesisPosPub struct {
	streamName string
	// TODO: add AWS credentials as optional
}

// KinesisPublisher docs here
func KinesisPublisher(streamName string) Publisher {
	return &kinesisPosPub{streamName: streamName}
}

// PublishPos docs here
// TODO: deal with buffering and batch writing
func (p *kinesisPosPub) Publish(bs []byte) error {
	return errors.New("Not implemented")
}
