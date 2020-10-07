package routesim

import (
	"errors"
	"math/rand"
	"testing"

	"github.com/golang/geo/s2"
	"github.com/gpontesss/routesim/pkg/gps"
	"github.com/gpontesss/routesim/pkg/gps/gpstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestingEmitter creates a FreqEmitter that will yield a list of
// positions imediatelly.
func TestingEmitter(id string, posl ...s2.LatLng) *FreqEmitter {
	return FreqEmitterWithTicker(
		gpstest.TestGPS(id, posl...),
		Ticker(len(posl)),
	)
}

func RandomIntRange(min int, max int) int {
	return rand.Intn(max-min) + min
}

func RandomLatLngs(len int) []s2.LatLng {
	lls := make([]s2.LatLng, len)
	for i := 0; i < len; i++ {
		lls[i] = s2.LatLngFromDegrees(
			float64(RandomIntRange(-90, 90)),
			float64(RandomIntRange(-180, 180)),
		)
	}
	return lls
}

// Returns a publisher that will fail at call n
func testingPublisher(n int) *testPosPub {
	pub := new(testPosPub)
	pub.failAt = n
	return pub
}

type testPosPub struct {
	mock.Mock
	failAt int
}

func (pub *testPosPub) PublishPos(pos gps.Position) error {
	args := pub.Called(pos.GPS.ID(), len(pub.Calls))
	if len(pub.Calls) == pub.failAt {
		return errors.New("Failed to publish position")
	}
	return args.Error(0)
}

func TestRouteSim(t *testing.T) {
	pub := testingPublisher(10)
	pub.On("PublishPos",
		mock.AnythingOfType("string"),
		mock.AnythingOfType("int")).
		Return(nil)

	emts := []*FreqEmitter{
		TestingEmitter("TEST0987", RandomLatLngs(5)...),
		TestingEmitter("TEST1234", RandomLatLngs(5)...),
	}

	sim := NewRouteSim(emts, pub)

	err := sim.Run()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to publish position")

	pub.AssertNumberOfCalls(t, "PublishPos", 10)
	pub.AssertCalled(t, "PublishPos", "TEST0987", mock.AnythingOfType("int"))
	pub.AssertCalled(t, "PublishPos", "TEST1234", mock.AnythingOfType("int"))
}
