package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang/geo/s2"
	"github.com/gpontesss/routesim/cmd/routesim/internal/routesim"
	"github.com/gpontesss/routesim/pkg/emitter"
	"github.com/gpontesss/routesim/pkg/gps"
	"github.com/gpontesss/routesim/pkg/pospub"
	"github.com/jonas-p/go-shp"
	"github.com/sirupsen/logrus"
)

type Config struct {
	GPSCfgArray     []GPSCofig      `json:"gps"`
	PublisherConfig PublisherConfig `json:"publisher"`
}

func (cfg Config) BuildRouteSim() (*routesim.RouteSim, error) {
	ems := make([]*emitter.PosEmitter, 0, len(cfg.GPSCfgArray))
	for _, gpsCfg := range cfg.GPSCfgArray {
		em, err := gpsCfg.BuildPosEmitter()
		if err != nil {
			return nil, fmt.Errorf("Error building PosEmitter: %w", err)
		}
		ems = append(ems, em)
	}

	pub, err := cfg.PublisherConfig.BuildPublisher()
	if err != nil {
		return nil, fmt.Errorf("Error building Publisher: %w", err)
	}
	return routesim.NewRouteSim(ems, pub), nil
}

type GPSCofig struct {
	// Relative path for shapefile describing GPS's route
	ShapefilePath string `json:"shapefile"`
	// Route mode that describes the behavior of the route when it reaches the
	// geometry's end
	Mode ModeGPS `json:"mode"`
	// Frequency in seconds that new positions should be sent
	Frequency Frequency `json:"frequency"`
	// GPS's distance rate of change (m/s)
	Velocity float64 `json:"velocity"`
	// Properties to append to GeoJSON
	Properties map[string]interface{} `json:"properties"`
}

func (cfg GPSCofig) BuildPosEmitter() (*emitter.PosEmitter, error) {
	gps, err := cfg.BuildGPS()
	if err != nil {
		return nil, fmt.Errorf("Error building GPS: %w", err)
	}
	return emitter.GPSPosEmitter(
		gps,
		time.Duration(cfg.Frequency),
		cfg.Properties,
	), nil
}

func (cfg GPSCofig) BuildGPS() (gps.GPS, error) {
	shprdr, err := shp.Open(cfg.ShapefilePath)
	if err != nil {
		return nil, err
	}

	path, err := s2PolylineFromShpReader(shprdr)
	if err != nil {
		return nil, fmt.Errorf("Error reading '%s': %w", cfg.ShapefilePath, err)
	}

	lw, err := BuildLineWalker(cfg.Mode, path)
	if err != nil {
		return nil, fmt.Errorf("Error building line walker: %w", err)
	}

	// TODO: define ID generation
	return gps.NewSimGPS("", cfg.Velocity, lw), nil
}

func s2PolylineFromShpReader(rdr *shp.Reader) (*s2.Polyline, error) {
	defer rdr.Close()

	if rdr.GeometryType != shp.POLYLINE {
		return nil, errors.New("Geometry type must be POLYLINE")
	}

	// The first geometry is the chosen
	rdr.Next()

	_, shape := rdr.Shape()
	pl := shape.(*shp.PolyLine)

	coords := make([]s2.LatLng, len(pl.Points))
	for i, pt := range pl.Points {
		coords[i] = s2.LatLngFromDegrees(pt.X, pt.Y)
	}

	return s2.PolylineFromLatLngs(coords), nil
}

func BuildLineWalker(mode ModeGPS, path *s2.Polyline) (gps.LineWalker, error) {
	switch mode {
	case BackAndForthMode:
		return gps.BackForthWalker(path), nil
	case RestartMode:
		return gps.RestartWalker(path), nil
	default:
		return nil, fmt.Errorf("Unknown GPS mode '%s'", mode)
	}
}

type ModeGPS string

const (
	BackAndForthMode ModeGPS = "BackAndForth"
	RestartMode              = "Restart"
)

func (m *ModeGPS) UnmarshalJSON(v []byte) error {
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return err
	}

	switch strings.ToLower(s) {
	case "backandforth":
		*m = BackAndForthMode
	case "restart":
		*m = RestartMode
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

func (cfg PublisherConfig) BuildPublisher() (pospub.PosPublisher, error) {
	switch cfg.Type {
	case KinesisPublisher:
		var kcfg kinesisPubCfg
		if err := json.Unmarshal(cfg.Options, &kcfg); err != nil {
			return nil, err
		}
		return pospub.KinesisPosPublisher(kcfg.StreamName), nil
	case LogPublisher:
		var lcfg logPubCfg
		if err := json.Unmarshal(cfg.Options, &lcfg); err != nil {
			return nil, err
		}
		return pospub.LogPosPub(logrus.New(), lcfg.Level), nil
	case ShpfilePublisher:
		var shpcfg shpPubCfg
		if err := json.Unmarshal(cfg.Options, &shpcfg); err != nil {
			return nil, err
		}
		return pospub.ShpfilePosPublisher(shpcfg.FilePath, shpcfg.Count)
	default:
		return nil, errors.New("Unkonwn publisher type")
	}
}

const (
	LogPublisher     PublisherType = "Log"
	KinesisPublisher               = "Kinesis"
	ShpfilePublisher               = "Shpfile"
)

type PublisherType string

func (t *PublisherType) UnmarshalJSON(v []byte) error {
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return err
	}

	switch strings.ToLower(s) {
	case "kinesis":
		*t = KinesisPublisher
	case "log":
		*t = LogPublisher
	case "shpfile":
		*t = ShpfilePublisher
	default:
		return fmt.Errorf("Unknown publisher type '%s'", s)
	}
	return nil
}

type kinesisPubCfg struct {
	StreamName string `json:"stream"`
}

type logPubCfg struct {
	Level logrus.Level `json:"level"`
}

type shpPubCfg struct {
	Count    int32  `json:"count"`
	FilePath string `json:"path"`
}

type Frequency time.Duration

func (f *Frequency) UnmarshalJSON(v []byte) error {
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return err
	}

	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*f = Frequency(dur)
	return nil
}
