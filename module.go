package linearactuatorwithposition

import (
	"context"
	"errors"
	"fmt"
	"time"

	gantry "go.viam.com/rdk/components/gantry"
	"go.viam.com/rdk/components/motor"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/referenceframe"
	"go.viam.com/rdk/resource"
)

var (
	LinearActuator   = resource.NewModel("gmulz", "linear-actuator-with-position", "linear-actuator")
	errUnimplemented = errors.New("unimplemented")
)

func init() {
	resource.RegisterComponent(gantry.API, LinearActuator,
		resource.Registration[gantry.Gantry, *Config]{
			Constructor: newLinearActuatorWithPositionLinearActuator,
		},
	)
}

type SensorConfig struct {
	Name          string `json:"name"`
	PositionField string `json:"position_field"`
}

type Config struct {
	StrokeLength     int32        `json:"stroke_length"`
	MaxExtensionTime *int32       `json:"max_extension_time,omitempty"`
	Motor            string       `json:"motor"`
	Sensor           SensorConfig `json:"position_sensor"`
}

// Validate ensures all parts of the config are valid and important fields exist.
// Returns implicit required (first return) and optional (second return) dependencies based on the config.
// The path is the JSON path in your robot's config (not the `Config` struct) to the
// resource being validated; e.g. "components.0".
func (cfg *Config) Validate(path string) ([]string, []string, error) {
	// Add config validation code here
	if cfg.StrokeLength <= 0 {
		return nil, nil, fmt.Errorf("stroke_length must be greater than 0")
	}
	if cfg.Motor == "" {
		return nil, nil, fmt.Errorf("motor must be specified")
	}
	if cfg.Sensor.Name == "" {
		return nil, nil, fmt.Errorf("position-sensor name must be specified")
	}
	if cfg.Sensor.PositionField == "" {
		return nil, nil, fmt.Errorf("position-sensor's position_field must be specified")
	}

	return []string{cfg.Motor, cfg.Sensor.Name}, nil, nil
}

type linearActuatorWithPositionLinearActuator struct {
	resource.AlwaysRebuild

	name resource.Name

	logger logging.Logger
	cfg    *Config

	cancelCtx  context.Context
	cancelFunc func()

	strokeLength        int32
	maxExtensionTime    *int32
	positionSensorField string
	positionSensor      sensor.Sensor
	motor               motor.Motor
	maxSensorPosition   *float64
	minSensorPosition   *float64
}

func newLinearActuatorWithPositionLinearActuator(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (gantry.Gantry, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewLinearActuator(ctx, deps, rawConf.ResourceName(), conf, logger)

}

func NewLinearActuator(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger) (gantry.Gantry, error) {

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	motor, err := motor.FromProvider(deps, conf.Motor)
	if err != nil {
		cancelFunc()
		return nil, err
	}
	positionSensor, err := sensor.FromProvider(deps, conf.Sensor.Name)
	if err != nil {
		cancelFunc()
		return nil, err
	}

	s := &linearActuatorWithPositionLinearActuator{
		name:                name,
		logger:              logger,
		cfg:                 conf,
		cancelCtx:           cancelCtx,
		cancelFunc:          cancelFunc,
		motor:               motor,
		positionSensor:      positionSensor,
		strokeLength:        conf.StrokeLength,
		maxExtensionTime:    conf.MaxExtensionTime,
		positionSensorField: conf.Sensor.PositionField,
	}
	return s, nil
}

func (s *linearActuatorWithPositionLinearActuator) Name() resource.Name {
	return s.name
}

// Position returns the position in meters.
func (s *linearActuatorWithPositionLinearActuator) Position(ctx context.Context, extra map[string]interface{}) ([]float64, error) {
	return nil, fmt.Errorf("not implemented")
}

// Lengths is the length of gantries in meters.
func (s *linearActuatorWithPositionLinearActuator) Lengths(ctx context.Context, extra map[string]interface{}) ([]float64, error) {
	return nil, fmt.Errorf("not implemented")
}

// Home runs the homing sequence of the gantry and returns true once completed.
func (s *linearActuatorWithPositionLinearActuator) Home(ctx context.Context, extra map[string]interface{}) (bool, error) {
	// Extend the actuator the full stroke length
	err := s.motor.SetPower(ctx, 1.0, extra)
	if err != nil {
		return false, fmt.Errorf("failed to start motor power extending actuator: %w", err)
	}
	if s.maxExtensionTime != nil {
		time.Sleep(time.Duration(*s.maxExtensionTime) * time.Second)
	} else {
		// todo: smarter way to do this is wait until the sensor values are somewhat stable
		time.Sleep(time.Duration(30) * time.Second)
	}
	err = s.motor.SetPower(ctx, 0.0, extra)
	if err != nil {
		return false, fmt.Errorf("failed to stop motor power extending actuator: %w", err)
	}

	// Read the position sensor
	readings, err := s.positionSensor.Readings(ctx, extra)
	if err != nil {
		return false, fmt.Errorf("failed to read position sensor at full extension: %w", err)
	}
	position, ok := readings[s.positionSensorField]
	if !ok {
		return false, fmt.Errorf("position sensor field %s not found", s.positionSensorField)
	}
	positionValue, ok := position.(float64)
	if !ok {
		return false, fmt.Errorf("position sensor field %s is not a float64", s.positionSensorField)
	}
	s.maxSensorPosition = &positionValue
	// todo: take several readings and take the average

	// Retract the actuator fully
	err = s.motor.SetPower(ctx, -1.0, extra)
	if err != nil {
		return false, fmt.Errorf("failed to start motor power retracting actuatory: %w", err)
	}

	if s.maxExtensionTime != nil {
		time.Sleep(time.Duration(*s.maxExtensionTime) * time.Second)
	} else {
		time.Sleep(time.Duration(30) * time.Second)
	}

	// Read the position sensor
	readings, err = s.positionSensor.Readings(ctx, extra)
	if err != nil {
		return false, fmt.Errorf("failed to read position sensor at retracted extension: %w", err)
	}
	positionValue, ok = position.(float64)
	if !ok {
		return false, fmt.Errorf("position sensor field %s is not a float64", s.positionSensorField)
	}
	s.minSensorPosition = &positionValue
	// todo: take several readings and take the average

	return true, nil
}

// MoveToPosition is in meters.
// This will block until done or a new operation cancels this one.
func (s *linearActuatorWithPositionLinearActuator) MoveToPosition(ctx context.Context, positionsMm []float64, speedsMmPerSec []float64, extra map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}

func (s *linearActuatorWithPositionLinearActuator) Stop(ctx context.Context, extra map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}

func (s *linearActuatorWithPositionLinearActuator) Kinematics(ctx context.Context) (referenceframe.Model, error) {
	var modelRetVal referenceframe.Model

	return modelRetVal, fmt.Errorf("not implemented")
}

func (s *linearActuatorWithPositionLinearActuator) CurrentInputs(ctx context.Context) ([]referenceframe.Input, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *linearActuatorWithPositionLinearActuator) GoToInputs(ctx context.Context, inputSteps ...[]referenceframe.Input) error {
	return fmt.Errorf("not implemented")
}

func (s *linearActuatorWithPositionLinearActuator) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *linearActuatorWithPositionLinearActuator) IsMoving(ctx context.Context) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (s *linearActuatorWithPositionLinearActuator) Close(context.Context) error {
	// Put close code here
	s.cancelFunc()
	return nil
}
