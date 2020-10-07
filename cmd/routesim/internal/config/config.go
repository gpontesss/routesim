package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang/geo/s2"
	"github.com/gpontesss/routesim/cmd/routesim/internal/routesim"
	"github.com/gpontesss/routesim/pkg/data"
	"github.com/gpontesss/routesim/pkg/gps"
	"github.com/jonas-p/go-shp"
)

// Config describes a JSON configuration for RouteSim
type Config struct {
	GPSCfgArray     []GPSConfig     `json:"gps"`
	PublisherConfig PublisherConfig `json:"publisher"`
}

// BuildRouteSim assembles a RouteSim
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

// GPSConfig describes a JSON configuration for a GPS and FreqEmitter
type GPSConfig struct {
	// Relative path for shapefile describing GPS's route
	ShapefilePath string `json:"shapefile"`
	// Route mode that describes the behavior of the route when it reaches the
	// geometry's end
	Mode WalkingModeGPS `json:"mode"`
	// Frequency in seconds that new positions should be sent
	Frequency Frequency `json:"frequency"`
	// GPS's distance rate of change (m/s)
	Velocity float64 `json:"velocity"`
	// Metadata to attach to the simulated device
	Metadata map[string]interface{} `json:"metadata"`
}

// BuildFreqEmitter assembles a FreqEmitter
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

// BuildGPS assembles a SimGPS
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

	return gps.NewSimGPS(cfg.Velocity, lw, cfg.Metadata), nil
}

// Gets a s2.Polyline from a shpfile reader
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

// BuildLineWalker assembles a LineWalker
func BuildLineWalker(mode WalkingModeGPS, path *s2.Polyline) (gps.LineWalker, error) {
	switch mode {
	case BackAndForthMode:
		return gps.BackForthWalker(path), nil
	case RestartMode:
		return gps.RestartWalker(path), nil
	default:
		return nil, fmt.Errorf("Unknown GPS mode '%s'", mode)
	}
}

// WalkingModeGPS identifies a GPS walking mode
type WalkingModeGPS string

const (
	// BackAndForthMode identifies a back-and-forth walking mode
	BackAndForthMode WalkingModeGPS = "BackAndForth"
	// RestartMode identifies a restart walking mode
	RestartMode = "Restart"
)

// UnmarshalJSON unmarshals a WalkingModeGPS
func (m *WalkingModeGPS) UnmarshalJSON(v []byte) error {
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

// PublisherConfig describes a JSON configuration for a position publisher
type PublisherConfig struct {
	// Publisher type name
	Type PublisherType `json:"type"`
	// Options specific for publisher
	Options json.RawMessage `json:"options"`
}

// BuildPublisher assembles a PosPublisher
func (cfg PublisherConfig) BuildPublisher() (data.PosPublisher, error) {
	switch cfg.Type {
	case KinesisPublisher:
		var kcfg kinesisPubCfg
		if err := json.Unmarshal(cfg.Options, &kcfg); err != nil {
			return nil, err
		}
		fmtr, err := kcfg.Format.GetFormatter()
		if err != nil {
			return nil, err
		}
		return data.PosFormatterPublisher(
			data.KinesisPublisher(kcfg.StreamName),
			fmtr,
		), nil

	case ShpfilePublisher:
		var shpcfg shpPubCfg
		if err := json.Unmarshal(cfg.Options, &shpcfg); err != nil {
			return nil, err
		}
		return data.ShpfilePublisher(shpcfg.FilePath, shpcfg.Count)

	case WebsocketPublisher:
		var wscfg wsCfg
		if err := json.Unmarshal(cfg.Options, &wscfg); err != nil {
			return nil, err
		}
		fmtr, err := wscfg.Format.GetFormatter()
		if err != nil {
			return nil, err
		}
		return data.PosFormatterPublisher(
			data.WebsocketPublisher(wscfg.Address, wscfg.Path),
			fmtr,
		), nil

	default:
		return nil, errors.New("Unkonwn publisher type")
	}
}

// PublisherType identifies the position publisher type
type PublisherType string

const (
	// KinesisPublisher identifies a Kinesis position publisher
	KinesisPublisher PublisherType = "Kinesis"
	// ShpfilePublisher identifies a Shpfile position publisher
	ShpfilePublisher = "Shpfile"
	// WebsocketPublisher identifies a Websocket position publisher
	WebsocketPublisher = "Websocket"
)

// UnmarshalJSON ummarshals a PublisherType
func (t *PublisherType) UnmarshalJSON(v []byte) error {
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return err
	}

	switch strings.ToLower(s) {
	case "kinesis":
		*t = KinesisPublisher
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
	StreamName string        `json:"stream"`
	Format     FormatterType `json:"format"`
}

type shpPubCfg struct {
	Count    int32  `json:"count"`
	FilePath string `json:"path"`
}

type wsCfg struct {
	Format  FormatterType `json:"format,omitempty"`
	Address string        `json:"address"`
	Path    string        `json:"path"`
}

// Frequency is the frequency that a GPS position should be emitted
type Frequency time.Duration

// UnmarshalJSON unmarshals a Frequency
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

// FormatterType identifies a PosFormatter type
type FormatterType string

const (
	// NotSpecifiedFormatter indentifies that a formatter wasn't specified
	NotSpecifiedFormatter FormatterType = ""
	// GeoJSONFormatter identifies a GeoJSON formatter
	GeoJSONFormatter = "GeoJSON"
)

// GetFormatter returns a PosFormatter instance according to its type
func (t FormatterType) GetFormatter() (data.PosFormatter, error) {
	switch t {
	case NotSpecifiedFormatter, GeoJSONFormatter:
		return data.GeoJSONFormatter, nil
	default:
		return nil, fmt.Errorf("Unknown formatter '%s'", t)
	}
}

// UnmarshalJSON unmarshals a FormatterType
func (t *FormatterType) UnmarshalJSON(v []byte) error {
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	case "geojson":
		*t = GeoJSONFormatter
	default:
		return fmt.Errorf("Unknown formatter type '%s'", s)
	}
	return nil
}
