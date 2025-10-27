package main

import (
	"linearactuatorwithposition"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
	gantry "go.viam.com/rdk/components/gantry"
)

func main() {
	// ModularMain can take multiple APIModel arguments, if your module implements multiple models.
	module.ModularMain(resource.APIModel{ gantry.API, linearactuatorwithposition.LinearActuator})
}
