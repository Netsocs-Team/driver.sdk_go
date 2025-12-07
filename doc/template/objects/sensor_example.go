package objects

import (
	"fmt"
	"log"
	"time"

	"github.com/Netsocs-Team/driver.sdk_go/pkg/objects"
)

// NewExampleSensor creates a temperature sensor object
//
// This example demonstrates how to create a sensor that measures temperature.
// Adapt this pattern for other sensor types: humidity, motion, smoke, water leak, etc.
func NewExampleSensor(objectID, deviceID string) objects.SensorObject {
	params := objects.NewSensorObjectParams{
		Metadata: objects.ObjectMetadata{
			ObjectID: objectID,
			Name:     "Example Temperature Sensor",
			Domain:   "temperature",
			DeviceID: deviceID,
			Tags:     []string{"temperature", "example", "indoor"},
			I18n: map[string]string{
				"es": "Sensor de Temperatura de Ejemplo",
				"en": "Example Temperature Sensor",
			},
		},

		// Setup function - called automatically after registration
		//
		// Use this to:
		// - Initialize the sensor
		// - Set sensor type and unit
		// - Set initial state and value
		// - Start background polling/updates
		SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
			log.Printf("Setting up sensor: %s", objectID)

			// Cast to SensorObject to access sensor-specific methods
			sensor := obj.(objects.SensorObject)

			// Set sensor type (Number, Text, Binary, or Battery)
			err := sensor.SetSensorType(objects.SensorObjectTypeNumber)
			if err != nil {
				return fmt.Errorf("failed to set sensor type: %w", err)
			}

			// Set unit of measurement
			err = sensor.SetUnitOfMeasurement("°C")
			if err != nil {
				return fmt.Errorf("failed to set unit: %w", err)
			}

			// Set initial state
			// Options: SENSOR_STATE_MEASUREMENT, SENSOR_STATE_TOTAL, SENSOR_STATE_TOTAL_INCREASING
			err = sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)
			if err != nil {
				return fmt.Errorf("failed to set state: %w", err)
			}

			// Set initial value
			err = sensor.SetValue("20.0")
			if err != nil {
				return fmt.Errorf("failed to set initial value: %w", err)
			}

			// Start periodic updates (optional)
			// In a real driver, replace this with actual device polling
			go periodicSensorUpdate(sensor)

			log.Printf("Sensor %s setup complete", objectID)
			return nil
		},

		// Optional: Handle bypass action (for alarm systems)
		//
		// This is called when a user bypasses the sensor (disables it temporarily)
		AlarmDetectorBypass: func(sensor objects.SensorObject, oc objects.ObjectController,
			payload objects.AlarmDetectorBypassPayload) (map[string]string, error) {

			log.Printf("Sensor %s bypassed", objectID)

			// TODO: Send bypass command to actual device
			// Example:
			// device, err := getDeviceConnection()
			// if err != nil {
			//     return nil, err
			// }
			// err = device.BypassSensor(objectID)
			// if err != nil {
			//     return nil, err
			// }

			// Update sensor state to indicate it's bypassed
			sensor.UpdateStateAttributes(map[string]string{
				"bypassed": "true",
			})

			return map[string]string{
				"status": "bypassed",
			}, nil
		},

		// Optional: Handle unbypass action
		//
		// This re-enables a bypassed sensor
		AlarmDetectorUnbypass: func(sensor objects.SensorObject, oc objects.ObjectController,
			payload objects.AlarmDetectorUnbypassPayload) (map[string]string, error) {

			log.Printf("Sensor %s unbypassed", objectID)

			// TODO: Send unbypass command to actual device
			// device, err := getDeviceConnection()
			// if err != nil {
			//     return nil, err
			// }
			// err = device.UnbypassSensor(objectID)

			// Update sensor state
			sensor.UpdateStateAttributes(map[string]string{
				"bypassed": "false",
			})

			return map[string]string{
				"status": "active",
			}, nil
		},

		// Optional: Handle custom actions
		//
		// Use this for sensor-specific operations not covered by standard actions
		CustomAction: func(sensor objects.SensorObject, oc objects.ObjectController,
			payload objects.CustomActionPayload) (map[string]string, error) {

			log.Printf("Custom action called on sensor %s", objectID)

			// TODO: Implement custom logic
			// Examples:
			// - Calibrate sensor
			// - Reset counters
			// - Change sensor sensitivity
			// - Perform self-test

			return map[string]string{
				"status": "completed",
			}, nil
		},
	}

	return objects.NewSensorObject(params)
}

// periodicSensorUpdate simulates reading sensor values from a device
//
// TODO: Replace this with actual device communication
//
// In a real driver, you would:
// 1. Connect to the device
// 2. Read the sensor value via API, Modbus, SNMP, etc.
// 3. Update the sensor value
// 4. Handle errors and reconnection
func periodicSensorUpdate(sensor objects.SensorObject) {
	// Simulate temperature readings every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	temperature := 20.0

	for range ticker.C {
		// Simulate temperature fluctuation
		// In real implementation: temperature = readFromDevice()
		temperature += (float64(time.Now().Unix()%3) - 1) * 0.5

		// Update sensor value
		err := sensor.SetValue(fmt.Sprintf("%.1f", temperature))
		if err != nil {
			log.Printf("Error updating sensor value: %v", err)
			continue
		}

		// Optionally update additional attributes
		err = sensor.UpdateStateAttributes(map[string]string{
			"last_updated": time.Now().Format(time.RFC3339),
			"battery_level": "85", // If sensor has battery
		})
		if err != nil {
			log.Printf("Error updating sensor attributes: %v", err)
		}

		log.Printf("Sensor updated: %.1f°C", temperature)
	}
}

// Examples of other sensor types:
//
// 1. Binary Sensor (motion, door, window):
//
// func NewMotionSensor(objectID, deviceID string) objects.SensorObject {
//     params := objects.NewSensorObjectParams{
//         Metadata: objects.ObjectMetadata{
//             ObjectID: objectID,
//             Name:     "Motion Sensor",
//             Domain:   "motion",
//             DeviceID: deviceID,
//         },
//         SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
//             sensor := obj.(objects.SensorObject)
//             sensor.SetSensorType(objects.SensorObjectTypeBinary)
//             sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)
//             sensor.SetValue("0") // 0 = no motion, 1 = motion detected
//             return nil
//         },
//     }
//     return objects.NewSensorObject(params)
// }
//
// 2. Counter Sensor (people counting, event count):
//
// func NewPeopleCounterSensor(objectID, deviceID string) objects.SensorObject {
//     params := objects.NewSensorObjectParams{
//         Metadata: objects.ObjectMetadata{
//             ObjectID: objectID,
//             Name:     "People Counter",
//             Domain:   "counter",
//             DeviceID: deviceID,
//         },
//         SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
//             sensor := obj.(objects.SensorObject)
//             sensor.SetSensorType(objects.SensorObjectTypeNumber)
//             sensor.SetState(objects.SENSOR_STATE_TOTAL_INCREASING)
//             sensor.SetValue("0")
//             return nil
//         },
//     }
//     return objects.NewSensorObject(params)
// }
//
// Use sensor.Increment() to increase counter
// Use sensor.Decrement() to decrease counter
//
// 3. Battery Sensor:
//
// func NewBatterySensor(objectID, deviceID string) objects.SensorObject {
//     params := objects.NewSensorObjectParams{
//         Metadata: objects.ObjectMetadata{
//             ObjectID: objectID,
//             Name:     "Battery Level",
//             Domain:   "battery",
//             DeviceID: deviceID,
//         },
//         SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
//             sensor := obj.(objects.SensorObject)
//             sensor.SetSensorType(objects.SensorObjectTypeBattery)
//             sensor.SetUnitOfMeasurement("%")
//             sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)
//             sensor.SetValue("100")
//             return nil
//         },
//     }
//     return objects.NewSensorObject(params)
// }
