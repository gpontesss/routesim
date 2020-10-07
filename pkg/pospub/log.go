package pospub

import (
	"encoding/json"

	"github.com/gpontesss/routesim/pkg/gps"
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
func (p *logPosPub) PublishPos(pos gps.Position) error {
	bs, err := json.Marshal(pos)
	if err != nil {
		return err
	}
	p.logger.Log(p.lvl, string(bs))
	return nil
}
