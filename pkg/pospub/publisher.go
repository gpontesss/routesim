package pospub

import geojson "github.com/paulmach/go.geojson"

// PosPublisher publishes a GeoJSON position as a Feature (advantage of adding
// custom properties)
type PosPublisher interface {
	PublishPos(geojson.Feature) error
}
