package main

import (
	"context"
	"linearactuatorwithposition"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	gantry "go.viam.com/rdk/components/gantry"
)

func main() {
	err := realMain()
	if err != nil {
		panic(err)
	}
}

func realMain() error {
	ctx := context.Background()
	logger := logging.NewLogger("cli")

	deps := resource.Dependencies{}
	// can load these from a remote machine if you need

	cfg := linearactuatorwithposition.Config{}

	thing, err := linearactuatorwithposition.NewLinearActuator(ctx, deps, gantry.Named("foo"), &cfg, logger)
	if err != nil {
		return err
	}
	defer thing.Close(ctx)

	return nil
}
