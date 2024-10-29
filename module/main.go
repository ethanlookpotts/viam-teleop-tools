package main

import (
	"go.viam.com/rdk/components/movementsensor"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
)

var (
	ModuleFamily      = resource.NewModelFamily("viz-team", "teleop")
	EverythingModel   = ModuleFamily.WithModel("everything")
	GlobetrotterModel = ModuleFamily.WithModel("globetrotter")
)

func main() {
	resource.RegisterComponent(sensor.API, EverythingModel, resource.Registration[sensor.Sensor, *EverythingConfig]{
		Constructor: newEverything,
	})

	resource.RegisterComponent(movementsensor.API, GlobetrotterModel, resource.Registration[movementsensor.MovementSensor, *GlobetrotterConfig]{
		Constructor: newGlobetrotter,
	})

	module.ModularMain(
		resource.APIModel{
			API:   sensor.API,
			Model: EverythingModel,
		}, resource.APIModel{
			API:   movementsensor.API,
			Model: GlobetrotterModel,
		})
}
