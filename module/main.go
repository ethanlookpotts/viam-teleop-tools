package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
)

var (
	ModuleFamily    = resource.NewModelFamily("viz-team", "teleop")
	EverythingModel = ModuleFamily.WithModel("everything")
)

type Everything struct {
	resource.Named
	resource.TriviallyCloseable
	resource.TriviallyReconfigurable
	cfg *EverythingConfig
}

type EverythingConfig struct {
	ExtraReadingsData map[string]any `json:"extra_readings_data"`
}

func (c *EverythingConfig) Validate(path string) ([]string, error) {
	return nil, nil
}

func main() {
	resource.RegisterComponent(sensor.API, EverythingModel, resource.Registration[resource.Resource, *EverythingConfig]{
		Constructor: newEverything,
	})

	module.ModularMain("everything", resource.APIModel{
		API:   sensor.API,
		Model: EverythingModel,
	})
}
func newEverything(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (resource.Resource, error) {
	cfg, err := resource.NativeConfig[*EverythingConfig](conf)
	if err != nil {
		return nil, err
	}
	return &Everything{Named: conf.ResourceName().AsNamed(), cfg: cfg}, nil
}

func (e *Everything) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"raw_input": cmd, "time_unix": time.Now().Unix()}, nil
}

// Feel free to change this. I just played around in desmos until something looked interesting.
func RandomFunction(x float64) float64 {
	return 5*math.Sin(3*x) + 3*math.Cos(5*x) + 2*math.Tan(x/2) +
		0.5*math.Exp(-math.Abs(x)) + 1.5*math.Pow(x, 2)*math.Sin(x) +
		0.2*math.Log(math.Abs(x)+1) + 5
}

func TimeDependentNoise() float64 {
	now := time.Now()
	nanos := now.UnixNano()
	// repeat every 2 hours
	period := 1e9 * 1 * 60 * 60 * 2
	phase := math.Sin(float64(nanos) / period)
	return RandomFunction(phase)
}

func TimeDependentString() string {
	now := time.Now()
	minutes := now.Minute()
	if minutes%2 == 0 {
		return "even"
	}
	return "odd"
}
func TimeDependentType() any {
	now := time.Now()
	minutes := now.Minute()
	noise := TimeDependentNoise()
	if minutes%2 == 0 {
		return fmt.Sprint(noise)
	}
	return noise
}

func (e *Everything) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"null":                     nil,
		"extra_readings_data":      e.cfg.ExtraReadingsData,
		"noise":                    TimeDependentNoise(),
		"noise_as_string":          fmt.Sprint(TimeDependentNoise()),
		"noise_as_string_or_float": TimeDependentType(),
		"minute_parity":            TimeDependentString(),
		"time_unix":                time.Now().Unix(),
		"const": map[string]any{
			"int":            5,
			"int_zero":       0,
			"float":          5.1,
			"float_neg":      -5.1,
			"float_large":    1e100,
			"inf_pos":        math.Inf(1),
			"inf_neg":        math.Inf(-1),
			"NaN":            math.NaN(),
			"bool_true":      true,
			"bool_false":     false,
			"str":            "hello!",
			"str_with_emoji": "👋🌍",
			"array":          []any{5, 6, 7},
			"nested_map": map[string]any{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			"maps_in_array": []any{
				map[string]any{"a": 1, "b": 2, "c": 3},
				map[string]any{"a": 4, "b": 5, "c": 6},
			},
			"key.with.dots":   1,
			"key_with_emoji🌍": 2,
			"byte_slice":      []byte("Hello, World!"),
		},
	}, nil
}
