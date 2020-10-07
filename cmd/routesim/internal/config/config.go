package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	"github.com/gpontesss/routesim/cmd/routesim/internal/routesim"
	"github.com/gpontesss/routesim/pkg/gps"
	"github.com/gpontesss/routesim/pkg/pospub"
	"github.com/jonas-p/go-shp"
	"github.com/sirupsen/logrus"
)

// Config docs here
type Config struct {
	GPSCfgArray     []GPSConfig     `json:"gps"`
	PublisherConfig PublisherConfig `json:"publisher"`
}

// BuildRouteSim docs here
func (cfg Config) BuildRouteSim() (*routesim.RouteSim, error) {
	emts := make([]*gps.FreqEmitter, 0, len(cfg.GPSCfgArray))
	for _, gpsCfg := range cfg.GPSCfgArray {
		emt, err := gpsCfg.BuildFreqEmitter()
		if err != nil {
			return nil, fmt.Errorf("Error building PosEmitter: %w", err)
		}
		emts = append(emts, emt)
	}

	pub, err := cfg.PublisherConfig.BuildPublisher()
	if err != nil {
		return nil, fmt.Errorf("Error building Publisher: %w", err)
	}
	return routesim.NewRouteSim(emts, pub), nil
}

// GPSConfig docs here
type GPSConfig struct {
	// Relative path for shapefile describing GPS's route
	ShapefilePath string `json:"shapefile"`
	// Route mode that describes the behavior of the route when it reaches the
	// geometry's end
	Mode ModeGPS `json:"mode"`
	// Frequency in seconds that new positions should be sent
	Frequency Frequency `json:"frequency"`
	// GPS's distance rate of change (m/s)
	Velocity float64 `json:"velocity"`
}

// BuildFreqEmitter docs here
func (cfg GPSConfig) BuildFreqEmitter() (*gps.FreqEmitter, error) {
	sgps, err := cfg.BuildGPS()
	if err != nil {
		return nil, fmt.Errorf("Error building GPS: %w", err)
	}
	return gps.NewFreqEmitter(
		sgps,
		time.Duration(cfg.Frequency),
	), nil
}

// BuildGPS docs here
func (cfg GPSConfig) BuildGPS() (gps.GPS, error) {
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

	return gps.NewSimGPS(uuid.New().String(), cfg.Velocity, lw), nil
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

// BuildLineWalker docs here
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

// ModeGPS docs here
type ModeGPS string

const (
	// BackAndForthMode  docs here
	BackAndForthMode ModeGPS = "BackAndForth"
	// RestartMode docs here
	RestartMode = "Restart"
)

// UnmarshalJSON docs here
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

// PublisherConfig docs here
type PublisherConfig struct {
	// Publisher type name
	Type PublisherType `json:"type"`
	// Options specific for publisher
	Options json.RawMessage `json:"options"`
}

// BuildPublisher docs here
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
	case WebsocketPublisher:
		var wscfg wsCfg
		if err := json.Unmarshal(cfg.Options, &wscfg); err != nil {
			return nil, err
		}
		return pospub.WebsocketPosPublisher(wscfg.Address, wscfg.Path), nil
	default:
		return nil, errors.New("Unkonwn publisher type")
	}
}

const (
	// LogPublisher docs here
	LogPublisher PublisherType = "Log"
	// KinesisPublisher docs here
	KinesisPublisher = "Kinesis"
	// ShpfilePublisher docs here
	ShpfilePublisher = "Shpfile"
	// WebsocketPublisher docs here
	WebsocketPublisher = "Websocket"
)

// PublisherType docs here
type PublisherType string

// UnmarshalJSON docs here
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
	case "websocket":
		*t = WebsocketPublisher
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

type wsCfg struct {
	Address string `json:"address"`
	Path    string `json:"path"`
}

// Frequency docs here
type Frequency time.Duration

// UnmarshalJSON docs here
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
