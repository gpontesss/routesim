package pospub

import (
	"errors"

	"github.com/jonas-p/go-shp"
	geojson "github.com/paulmach/go.geojson"
)

type shpfilePosPub struct {
	wtr   *shp.Writer
	curri int32
	stopi int32
}

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

func (p *shpfilePosPub) PublishPos(pos geojson.Feature) error {
	pt := pos.Geometry.Point
	coord := &shp.Point{
		X: pt[0],
		Y: pt[1],
	}
	p.curri = p.wtr.Write(coord)

	if p.curri+1 == p.stopi {
		defer p.wtr.Close()
		return errors.New("Reached desired positions count")
	}

	return nil
}
