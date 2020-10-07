package data

import (
	"encoding/json"

	"github.com/gpontesss/routesim/pkg/gps"
	geojson "github.com/paulmach/go.geojson"
)

// PosFormatter describes a position formatter. It outputs a byte stream, for
// a more friendly/agostic interaction with data publishers.
type PosFormatter interface {
	// Formats a position to a byte stream
	Format(gps.Position) ([]byte, error)
}

// PosFormatterFunc is a helper to transform functions to Formatters
type PosFormatterFunc func(gps.Position) ([]byte, error)

// Format formats a position into a byte stream
func (f PosFormatterFunc) Format(pos gps.Position) ([]byte, error) {
	return f(pos)
}

var (
	// GeoJSONFormatter formats a position into a GeoJSON Feature point
	GeoJSONFormatter = geoJSONFormatter()
)

func geoJSONFormatter() PosFormatter {
	return PosFormatterFunc(func(pos gps.Position) ([]byte, error) {
		return json.Marshal(geojson.Feature{
			Type:       "Feature",
			ID:         pos.GPS.ID(),
			Properties: pos.GPS.Metadata(),
			Geometry: &geojson.Geometry{
				Type: geojson.GeometryPoint,
				Point: []float64{
					pos.Lat.Degrees(),
					pos.Lng.Degrees(),
				},
			},
		})
	})
}
