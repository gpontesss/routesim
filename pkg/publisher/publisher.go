package publisher

import geojson "github.com/paulmach/go.geojson"

// Publisher publishes a GeoJSON position
type Publisher interface {
	PublishPos(geojson.Feature) error
}
