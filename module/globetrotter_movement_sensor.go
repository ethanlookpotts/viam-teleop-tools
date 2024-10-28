package main

import (
	"context"
	"math"
	"math/rand/v2"
	"time"

	"github.com/golang/geo/r3"
	geo "github.com/kellydunn/golang-geo"
	"go.viam.com/rdk/components/movementsensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/spatialmath"
)

type Globetrotter struct {
	resource.Named
	resource.TriviallyCloseable
	resource.TriviallyReconfigurable
	cfg *GlobetrotterConfig
}

var (
	path = []*geo.Point{
		// nyc
		geo.NewPoint(40.66, -73.94),
		// glasgow
		geo.NewPoint(55.86, -4.25),
		// tip of brasil
		geo.NewPoint(-6.95, -34.89),
	}
	noiseRadiusM = float64(20)
	// 1 day
	cycleDuration = int64(60 * 60 * 24)
)

// Not sure why this wasn't packaged with the geo funcs.
// http://www.movable-type.co.uk/scripts/latlong.html
func greatCirclePoint(p1, p2 *geo.Point, f float64) *geo.Point {
	// Convert lat/lng to radians
	lat1 := p1.Lat() * math.Pi / 180.0
	lng1 := p1.Lng() * math.Pi / 180.0
	lat2 := p2.Lat() * math.Pi / 180.0
	lng2 := p2.Lng() * math.Pi / 180.0

	angularDistance := p1.GreatCircleDistance(p2)
	delta := angularDistance / geo.EARTH_RADIUS

	a := math.Sin((1-f)*delta) / math.Sin(delta)
	b := math.Sin(f*delta) / math.Sin(delta)

	x := a*math.Cos(lat1)*math.Cos(lng1) + b*math.Cos(lat2)*math.Cos(lng2)
	y := a*math.Cos(lat1)*math.Sin(lng1) + b*math.Cos(lat2)*math.Sin(lng2)
	z := a*math.Sin(lat1) + b*math.Sin(lat2)

	lat := math.Atan2(z, math.Sqrt(x*x+y*y))
	lng := math.Atan2(y, x)

	// Convert back to degrees
	return geo.NewPoint(lat*180.0/math.Pi, lng*180.0/math.Pi)
}

type Position struct {
	Point    *geo.Point
	Altitude float64
	Compass  float64
}

func position(time time.Time) *Position {
	cycleSpot := float64(time.Unix()%cycleDuration) / float64(cycleDuration)
	// one cycle per day. Time between points is equally spaced
	// (len(path)) instead of (len(path)-1) because we want to loop back to the start
	pair := math.Floor(cycleSpot * float64(len(path)))

	// percent between pair
	percentBetweenPair := cycleSpot*float64(len(path)) - pair

	p1 := path[int(pair)]
	p2 := path[(int(pair)+1)%len(path)]
	point := greatCirclePoint(p1, p2, percentBetweenPair)
	compass := point.BearingTo(p2)
	pointWithNoise := geo.NewPoint(
		point.Lat()+((rand.Float64()*2-1)*noiseRadiusM)/geo.EARTH_RADIUS,
		point.Lng()+((rand.Float64()*2-1)*noiseRadiusM)/geo.EARTH_RADIUS,
	)

	return &Position{
		Point:    pointWithNoise,
		Altitude: 1,
		Compass:  compass,
	}
}

type GlobetrotterConfig struct {
	ExtraReadingsData map[string]any `json:"extra_readings_data"`
}

func (c *GlobetrotterConfig) Validate(path string) ([]string, error) {
	return nil, nil
}

func newGlobetrotter(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (movementsensor.MovementSensor, error) {
	cfg, err := resource.NativeConfig[*GlobetrotterConfig](conf)
	if err != nil {
		return nil, err
	}
	return &Globetrotter{Named: conf.ResourceName().AsNamed(), cfg: cfg}, nil
}

func (e *Globetrotter) DoCommand(ctx context.Context, cmd map[string]any) (map[string]any, error) {
	return map[string]any{"raw_input": cmd, "time_unix": time.Now().Unix()}, nil
}

// Position returns the current GeoPoint (latitude, longitude) and altitude of the movement sensor above sea level in meters.
// Supported by GPS models.
func (e *Globetrotter) Position(ctx context.Context, extra map[string]any) (*geo.Point, float64, error) {
	pos := position(time.Now())
	return pos.Point, pos.Altitude, nil
}

// LinearVelocity returns the current linear velocity as a 3D vector in meters per second.
func (e *Globetrotter) LinearVelocity(ctx context.Context, extra map[string]any) (r3.Vector, error) {
	return r3.Vector{}, nil
}

// AngularVelcoity returns the current angular velocity as a 3D vector in degrees per second.
func (e *Globetrotter) AngularVelocity(ctx context.Context, extra map[string]any) (spatialmath.AngularVelocity, error) {
	return spatialmath.AngularVelocity{}, nil
}

// LinearAcceleration returns the current linear acceleration as a 3D vector in meters per second per second.
func (e *Globetrotter) LinearAcceleration(ctx context.Context, extra map[string]any) (r3.Vector, error) {
	return r3.Vector{}, nil
}

// CompassHeading returns the current compass heading in degrees.
func (e *Globetrotter) CompassHeading(ctx context.Context, extra map[string]any) (float64, error) {
	pos := position(time.Now())
	return pos.Compass, nil
}

// Orientation returns the current orientation of the movement sensor.
func (e *Globetrotter) Orientation(ctx context.Context, extra map[string]any) (spatialmath.Orientation, error) {
	return spatialmath.NewZeroOrientation(), nil
}

// Properties returns the supported properties of the movement sensor.
func (e *Globetrotter) Properties(ctx context.Context, extra map[string]any) (*movementsensor.Properties, error) {
	return &movementsensor.Properties{
		PositionSupported:       true,
		CompassHeadingSupported: true,
	}, nil
}

// Accuracy returns the reliability metrics of the movement sensor,
// including various parameters to access the sensor's accuracy and precision in different dimensions.
func (e *Globetrotter) Accuracy(ctx context.Context, extra map[string]any) (*movementsensor.Accuracy, error) {
	return nil, nil
}

func (e *Globetrotter) Readings(ctx context.Context, extra map[string]any) (map[string]any, error) {
	return map[string]any{
		"pos": "ocean",
	}, nil
}
