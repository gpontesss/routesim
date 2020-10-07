package data

import "github.com/gpontesss/routesim/pkg/gps"

// Publisher publishes data to a resource
type Publisher interface {
	Publish([]byte) error
}

// PosPublisher publishes a position to a resource
type PosPublisher interface {
	PublishPos(gps.Position) error
}

// PosPublisherFunc is a helper for turning functions into PosPublishers
type PosPublisherFunc func(gps.Position) error

// PublishPos publishes a position to a resource
func (f PosPublisherFunc) PublishPos(pos gps.Position) error {
	return f(pos)
}

// PosFormatterPublisher returns a PosPublisher data applies a Formatter to a
// position and publishes the data to a Publisher
func PosFormatterPublisher(pub Publisher, fmtr PosFormatter) PosPublisher {
	return PosPublisherFunc(func(pos gps.Position) error {
		bs, err := fmtr.Format(pos)
		if err != nil {
			return err
		}
		return pub.Publish(bs)
	})
}
