package main

import (
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
)

var (
	ModuleFamily    = resource.NewModelFamily("viz-team", "teleop")
	EverythingModel = ModuleFamily.WithModel("everything")
)

func main() {
	resource.RegisterComponent(sensor.API, EverythingModel, resource.Registration[resource.Resource, *EverythingConfig]{
		Constructor: newEverything,
	})

	module.ModularMain("everything", resource.APIModel{
		API:   sensor.API,
		Model: EverythingModel,
	})
}
