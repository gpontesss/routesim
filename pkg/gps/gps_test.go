package gps

import (
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TimeFunc returns a function that time objects sequentially at each call
func TimeFunc(rets ...time.Time) func() time.Time {
	i := -1
	return func() time.Time {
		i++
		return rets[i]
	}
}

type fakeLineWalker struct {
	mock.Mock
}

func (f *fakeLineWalker) Walk(dist Distance) (s2.LatLng, bool) {
	args := f.Called(dist)
	return args.Get(0).(s2.LatLng), args.Bool(1)
}

func (f *fakeLineWalker) Reset() {}

func TestSimGPS(t *testing.T) {
	now := time.Now()
	nowFunc = TimeFunc(now, now.Add(5*time.Second), now.Add(8*time.Second))

	flw := new(fakeLineWalker)
	flw.On("Walk", DistanceFromMeters(50)).
		Return(s2.LatLngFromDegrees(45, 45), false)
	flw.On("Walk", DistanceFromMeters(30)).
		Return(s2.LatLngFromDegrees(90, 0), false)

	metadata := map[string]interface{}{
		"vehicle": "car",
	}
	gps := NewSimGPS(10, flw, metadata)

	assert.NotEqual(t, gps.ID(), "")
	assert.Equal(t, s2.LatLngFromDegrees(45, 45), gps.CurrentPos().LatLng)
	assert.Equal(t, s2.LatLngFromDegrees(90, 0), gps.CurrentPos().LatLng)
	assert.Equal(t, "car", gps.Metadata()["vehicle"])

	flw.AssertNumberOfCalls(t, "Walk", 2)
	flw.AssertExpectations(t)
}
