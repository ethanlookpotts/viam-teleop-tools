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
	PetModel          = ModuleFamily.WithModel("pet")
)

func main() {
	resource.RegisterComponent(sensor.API, EverythingModel, resource.Registration[sensor.Sensor, *EverythingConfig]{
		Constructor: newEverything,
	})

	resource.RegisterComponent(movementsensor.API, GlobetrotterModel, resource.Registration[movementsensor.MovementSensor, *GlobetrotterConfig]{
		Constructor: newGlobetrotter,
	})

	resource.RegisterComponent(sensor.API, PetModel, resource.Registration[sensor.Sensor, *PetConfig]{
		Constructor: newPet,
	})

	module.ModularMain(
		resource.APIModel{
			API:   sensor.API,
			Model: EverythingModel,
		}, resource.APIModel{
			API:   movementsensor.API,
			Model: GlobetrotterModel,
		}, resource.APIModel{
			API:   sensor.API,
			Model: PetModel,
		})
}
