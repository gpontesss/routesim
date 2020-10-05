package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang/geo/s2"
	"github.com/gpontesss/routesim/pkg/gps"
	"github.com/gpontesss/routesim/pkg/publisher"
	"github.com/jonas-p/go-shp"
)

type Config struct {
	GPSCfgArray     []GPSCofig      `json:"gps"`
	PublisherConfig PublisherConfig `json:"publisher"`
}

type GPSCofig struct {
	// Relative path for shapefile describing GPS's route
	ShapefilePath string `json:"shapefile"`
	// Route mode that describes the behavior of the route when it reaches the
	// geometry's end
	Mode ModeGPS `json:"mode"`
	// Frequency in seconds that new positions should be sent
	Frequency time.Duration `json:"frequency"`
	// GPS's distance rate of change (km/h)
	Velocity float64 `json:"velocity"`
	// Properties to append to GeoJSON
	Properties map[string]interface{}
}

func (cfg *GPSCofig) BuildGPS() (gps.GPS, error) {
	shpfile, err := shp.Open(cfg.ShapefilePath)
	if err != nil {
		return nil, err
	}

	if shpfile.GeometryType != shp.POLYLINE {
		return nil, fmt.Errorf("Error reading '%s': geometry type must be POLYLINE", cfg.ShapefilePath)
	}

	// The first geometry is the chosen
	shpfile.Next()

	_, shape := shpfile.Shape()
	pl := shape.(*shp.PolyLine)

	coords := make([]s2.LatLng, 0, len(pl.Points))
	for i, pt := range pl.Points {
		coords[i] = s2.LatLngFromDegrees(pt.X, pt.Y)
	}

	path := s2.PolylineFromLatLngs(coords)
	_ = path

	panic("Unimplemented")
}

type ModeGPS string

const (
	BackAndForthMode ModeGPS = "BackAndForth"
	Restart                  = "Restart"
)

func (m ModeGPS) UnmarshalJSON(v []byte) error {
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return err
	}

	switch strings.ToLower(s) {
	case "backandforth":
		m = BackAndForthMode
	case "restart":
		m = RestartMode
	default:
		return fmt.Errorf("Unknown mode '%s'", s)
	}
	return nil
}

type PublisherConfig struct {
	// Publisher type name
	Type PublisherType `json:"type"`
	// Options specific for publisher
	Options json.RawMessage `json:"options"`
}

func (cfg PublisherConfig) BuildPublisher() (publisher.Publisher, error) {
	switch cfg.Type {
	case KinesisPublisher:
		var kcfg kinesisCfg
		if err := json.Unmarshal(cfg.Options, &kcfg); err != nil {
			return nil, err
		}
		return publisher.NewKinesisPublisher(kcfg.StreamName), nil
	default:
		return nil, errors.New("Unkonwn driver type")
	}
}

const (
	KinesisPublisher PublisherType = "Kinesis"
)

type PublisherType string

func (t PublisherType) UnmarshalJSON(v []byte) error {
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return err
	}

	switch strings.ToLower(s) {
	case "kinesis":
		t = KinesisPublisher
	default:
		return fmt.Errorf("Unknown publisher '%s'", s)
	}
	return nil
}

type kinesisCfg struct {
	StreamName string `json:"stream"`
}
