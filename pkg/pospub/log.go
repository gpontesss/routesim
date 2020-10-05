package pospub

import (
	"encoding/json"
	"fmt"

	geojson "github.com/paulmach/go.geojson"
	"github.com/sirupsen/logrus"
)

type logPosPub struct {
	logger *logrus.Logger
	lvl    logrus.Level
}

func LogPosPub(logger *logrus.Logger, lvl logrus.Level) PosPublisher {
	return &logPosPub{}
}

func (p *logPosPub) PublishPos(pos geojson.Feature) error {
	bs, err := json.Marshal(pos)
	if err != nil {
		return err
	}
	// p.logger.Log(p.lvl, string(bs))
	fmt.Printf("%s\n", bs)
	return nil
}
