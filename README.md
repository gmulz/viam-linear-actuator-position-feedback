# Module linear-actuator-with-position

A linear actuator module that uses an external position sensor for accurate positioning. This module implements the gantry API and provides position feedback through a separate sensor component.

## Model gmulz:linear-actuator-with-position:linear-actuator

This model implements a linear actuator with position feedback. It requires a motor for actuation and a sensor for position feedback. The actuator must be homed before use to establish position limits.

### Configuration

The following attribute template can be used to configure this model:

```json
{
  "stroke_length": <float>,
  "motor": "<string>",
  "position_sensor": {
    "name": "<string>",
    "position_field": "<string>"
  },
  "max_extension_time": <float>
}
```

#### Attributes

The following attributes are available for this model:

| Name                             | Type   | Inclusion | Description                                                                                     |
| -------------------------------- | ------ | --------- | ----------------------------------------------------------------------------------------------- |
| `stroke_length`                  | float  | Required  | The length of the actuator stroke in millimeters                                                |
| `motor`                          | string | Required  | The name of the motor component used to drive the actuator                                      |
| `position_sensor`                | object | Required  | The configuration for the position sensor, containing `name` and `position_field`               |
| `position_sensor.name`           | string | Required  | The name of the sensor component that provides position feedback                                |
| `position_sensor.position_field` | string | Required  | The field name in the sensor readings that contains the position value                          |
| `max_extension_time`             | float  | Optional  | Maximum time in seconds for full extension/retraction (defaults to 30 seconds if not specified) |

### Operation

Before using the actuator, you must run the `Home()` command. The homing sequence:

1. Extends the actuator to its maximum position
2. Reads the sensor value and stores it as the maximum position
3. Retracts the actuator to its minimum position
4. Reads the sensor value and stores it as the minimum position

After homing, the actuator maps sensor readings to actual positions based on these limits. It assumes the position sensor values are linear with the extension of the actuator

### Gantry API Methods

- **Position()**: Returns the current position in millimeters. Requires homing first.
- **Lengths()**: Returns the stroke length in millimeters.
- **Home()**: Calibrates the actuator by finding min/max positions.
- **MoveToPosition()**: Moves the actuator to a specified position (in millimeters). Blocks until complete or times out.
- **Stop()**: Stops the actuator immediately.
- **IsMoving()**: Checks if the actuator is currently moving.

### Example Configuration

```json
{
  "name": "my-linear-actuator",
  "api": "rdk:component:gantry",
  "model": "gmulz:linear-actuator-with-position:linear-actuator",
  "attributes": {
    "stroke_length": 1000.0,
    "motor": "actuator-motor",
    "position_sensor": {
      "name": "position-sensor",
      "position_field": "position"
    },
    "max_extension_time": 25.0
  }
}
```
