package pospub

import (
	"errors"

	"github.com/gpontesss/routesim/pkg/gps"
	"github.com/jonas-p/go-shp"
)

type shpfilePosPub struct {
	wtr   *shp.Writer
	curri int32
	stopi int32
}

// ShpfilePosPublisher docs here
func ShpfilePosPublisher(filePath string, count int32) (PosPublisher, error) {
	wtr, err := shp.Create(filePath, shp.POINT)
	if err != nil {
		return nil, err
	}

	return &shpfilePosPub{
		wtr:   wtr,
		curri: 0,
		stopi: count,
	}, nil
}

// PublishPos docs here
func (p *shpfilePosPub) PublishPos(pos gps.Position) error {
	coord := &shp.Point{
		X: pos.Lat.Degrees(),
		Y: pos.Lng.Degrees(),
	}
	p.curri = p.wtr.Write(coord)

	if p.curri+1 == p.stopi {
		defer p.wtr.Close()
		return errors.New("Reached desired positions count")
	}

	return nil
}
