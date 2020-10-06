package pospub

import (
	"encoding/json"

	geojson "github.com/paulmach/go.geojson"
	"github.com/sirupsen/logrus"
)

type logPosPub struct {
	logger *logrus.Logger
	lvl    logrus.Level
}

// LogPosPub docs here
func LogPosPub(logger *logrus.Logger, lvl logrus.Level) PosPublisher {
	return &logPosPub{}
}

// PublishPos docs here
func (p *logPosPub) PublishPos(pos geojson.Feature) error {
	bs, err := json.Marshal(pos)
	if err != nil {
		return err
	}
	p.logger.Log(p.lvl, string(bs))
	return nil
}
