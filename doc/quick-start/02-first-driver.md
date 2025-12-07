# Your First Driver

In this guide, you'll build a minimal but complete working driver that creates a temperature sensor and connects to the Netsocs platform.

## What You'll Build

A driver that:
- Connects to the Netsocs platform
- Creates a temperature sensor object
- Updates the sensor value
- Listens for configuration requests from the platform

**Time to complete**: 10-15 minutes

## Prerequisites

- Completed the [Installation](01-installation.md) guide
- Have your `driver.netsocs.json` configuration file ready

## Step 1: Create Your Main File

Create a new file `main.go` in your project root:

```go
package main

import (
    "log"
    "time"

    "github.com/Netsocs-Team/driver.sdk_go/pkg/client"
    "github.com/Netsocs-Team/driver.sdk_go/pkg/objects"
)

func main() {
    log.Println("Starting temperature sensor driver...")

    // Step 1: Initialize the client from driver.netsocs.json
    c, err := client.New()
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    log.Println("Client initialized successfully")

    // Step 2: Create a temperature sensor object
    sensor := createTemperatureSensor()

    // Step 3: Register the sensor with the platform
    err = c.RegisterObject(sensor)
    if err != nil {
        log.Fatalf("Failed to register sensor: %v", err)
    }
    log.Println("Sensor registered successfully")

    // Step 4: Simulate temperature updates (in a real driver, read from device)
    go simulateTemperatureUpdates(sensor)

    // Step 5: Listen for configuration requests (blocking call)
    log.Println("Driver ready, listening for requests...")
    err = c.ListenConfig()
    if err != nil {
        log.Fatalf("ListenConfig error: %v", err)
    }
}

// createTemperatureSensor creates and configures a temperature sensor object
func createTemperatureSensor() objects.SensorObject {
    params := objects.NewSensorObjectParams{
        // Define the sensor metadata
        Metadata: objects.ObjectMetadata{
            ObjectID: "temp_sensor_01",
            Name:     "Living Room Temperature",
            Domain:   "temperature",
            DeviceID: "demo_device_001",
            Tags:     []string{"temperature", "indoor"},
        },

        // Setup function - called automatically after registration
        SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
            log.Println("Sensor setup called")

            // Cast to SensorObject to access sensor-specific methods
            sensor := obj.(objects.SensorObject)

            // Set the sensor type to number (for numeric values)
            err := sensor.SetSensorType(objects.SensorObjectTypeNumber)
            if err != nil {
                return err
            }

            // Set the unit of measurement
            err = sensor.SetUnitOfMeasurement("°C")
            if err != nil {
                return err
            }

            // Set initial state and value
            err = sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)
            if err != nil {
                return err
            }

            err = sensor.SetValue("20.0")
            if err != nil {
                return err
            }

            log.Println("Sensor setup completed")
            return nil
        },
    }

    return objects.NewSensorObject(params)
}

// simulateTemperatureUpdates simulates reading temperature from a device
// In a real driver, replace this with actual device communication
func simulateTemperatureUpdates(sensor objects.SensorObject) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    temperature := 20.0

    for range ticker.C {
        // Simulate temperature fluctuation
        temperature += (float64(time.Now().Unix() % 3) - 1) * 0.5

        // Update the sensor value
        err := sensor.SetValue(fmt.Sprintf("%.1f", temperature))
        if err != nil {
            log.Printf("Error updating temperature: %v", err)
        } else {
            log.Printf("Temperature updated: %.1f°C", temperature)
        }
    }
}
```

**Important**: Add the `fmt` import:

```go
import (
    "fmt"
    "log"
    "time"

    "github.com/Netsocs-Team/driver.sdk_go/pkg/client"
    "github.com/Netsocs-Team/driver.sdk_go/pkg/objects"
)
```

## Step 2: Understand the Code

Let's break down what each part does:

### 1. Client Initialization

```go
c, err := client.New()
```

- Reads `driver.netsocs.json` from the current directory
- Initializes connection to the Netsocs platform
- Sets up WebSocket for receiving actions
- Returns a configured client instance

### 2. Sensor Creation

```go
params := objects.NewSensorObjectParams{
    Metadata: objects.ObjectMetadata{
        ObjectID: "temp_sensor_01",  // Unique ID within your driver
        Name:     "Living Room Temperature",  // Display name
        Domain:   "temperature",     // Logical grouping
        DeviceID: "demo_device_001", // Parent device ID
        Tags:     []string{"temperature", "indoor"},  // Searchable tags
    },
    SetupFn: func(...) error { ... }
}
```

**Metadata fields:**
- `ObjectID`: Must be unique within your driver
- `Name`: Displayed in the UI
- `Domain`: Groups related objects (e.g., "temperature", "camera", "door")
- `DeviceID`: Links the object to a device
- `Tags`: Helps users find and filter objects

### 3. Setup Function

The `SetupFn` is called automatically after the object is registered. Use it to:

```go
SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
    sensor := obj.(objects.SensorObject)  // Cast to access sensor methods
    sensor.SetSensorType(objects.SensorObjectTypeNumber)  // Number, Text, Binary, or Battery
    sensor.SetUnitOfMeasurement("°C")     // Display unit
    sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)  // Initial state
    sensor.SetValue("20.0")               // Initial value
    return nil
}
```

### 4. Object Registration

```go
err = c.RegisterObject(sensor)
```

This:
- Sends the object metadata to the platform
- Calls the `SetupFn` automatically
- Makes the object visible in the UI
- Enables action handling for the object

### 5. Listen for Requests

```go
err = c.ListenConfig()
```

This is a **blocking call** that:
- Opens a WebSocket connection to the platform
- Listens for configuration requests
- Keeps the driver running
- Should be the last call in `main()`

## Step 3: Build and Run

Build your driver:

```bash
go build -o my-driver
```

Run it:

```bash
./my-driver
```

You should see output like:

```
2025/01/15 10:30:00 Starting temperature sensor driver...
2025/01/15 10:30:00 Client initialized successfully
2025/01/15 10:30:00 Sensor setup called
2025/01/15 10:30:00 Sensor setup completed
2025/01/15 10:30:00 Sensor registered successfully
2025/01/15 10:30:00 Driver ready, listening for requests...
2025/01/15 10:30:10 Temperature updated: 20.5°C
2025/01/15 10:30:20 Temperature updated: 20.0°C
```

## Step 4: Verify in the Platform

1. Log in to the Netsocs platform
2. Navigate to **Devices** or **Objects**
3. You should see your temperature sensor: "Living Room Temperature"
4. The current value and state should be visible
5. Watch the value update every 10 seconds

## What Happens Behind the Scenes

When you run your driver:

1. **Authentication**: SDK authenticates using `driver_key` and `token`
2. **Registration**: Sensor object is created in the platform database
3. **WebSocket Connection**: Driver establishes persistent connection
4. **State Updates**: Temperature updates sent via HTTP API
5. **Ready for Actions**: Driver listens for user commands from platform

```
Your Driver              Netsocs Platform
     │                          │
     │──── Auth Request ────────▶
     │◀─── Auth Success ─────────│
     │                          │
     │──── Register Sensor ─────▶
     │◀─── Object Created ───────│
     │                          │
     │═══ WebSocket Open ═══════│
     │                          │
     │──── Update State ────────▶
     │◀─── State Saved ──────────│
     │                          │
     │◀═══ Listen for Actions ═══│
```

## Extending the Example

### Add More Sensors

You can register multiple objects:

```go
sensor1 := createTemperatureSensor("temp_01", "Living Room")
sensor2 := createTemperatureSensor("temp_02", "Bedroom")

c.RegisterObject(sensor1)
c.RegisterObject(sensor2)
```

### Add Humidity Sensor

```go
func createHumiditySensor() objects.SensorObject {
    params := objects.NewSensorObjectParams{
        Metadata: objects.ObjectMetadata{
            ObjectID: "humidity_01",
            Name:     "Living Room Humidity",
            Domain:   "humidity",
            DeviceID: "demo_device_001",
        },
        SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
            sensor := obj.(objects.SensorObject)
            sensor.SetSensorType(objects.SensorObjectTypeNumber)
            sensor.SetUnitOfMeasurement("%")
            sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)
            sensor.SetValue("45.0")
            return nil
        },
    }
    return objects.NewSensorObject(params)
}
```

### Read from Actual Device

Replace `simulateTemperatureUpdates` with actual device communication:

```go
func readTemperatureFromDevice(sensor objects.SensorObject, deviceIP string) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        // Make HTTP request to your device
        resp, err := http.Get(fmt.Sprintf("http://%s/api/temperature", deviceIP))
        if err != nil {
            log.Printf("Error reading device: %v", err)
            continue
        }

        // Parse response
        var data struct {
            Temperature float64 `json:"temperature"`
        }
        json.NewDecoder(resp.Body).Decode(&data)
        resp.Body.Close()

        // Update sensor
        sensor.SetValue(fmt.Sprintf("%.1f", data.Temperature))
    }
}
```

## Common Issues

### "Failed to create client"

- Ensure `driver.netsocs.json` exists in the current directory
- Verify all required fields are present
- Check JSON syntax is valid

### "Failed to register sensor"

- Ensure `ObjectID` is unique
- Check network connectivity to platform
- Verify credentials are correct

### "Connection refused"

- Ensure `driver_hub_host` URL is correct
- Check firewall settings
- Verify platform is accessible

## Next Steps

Congratulations! You've created your first working driver. Next, learn about:

- [Understanding Objects](03-understanding-objects.md) - Deep dive into object types and lifecycle
- [Configuration Handlers](04-configuration-handlers.md) - Handle requests from the platform

## Complete Code

<details>
<summary>Click to see the complete working code</summary>

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/Netsocs-Team/driver.sdk_go/pkg/client"
    "github.com/Netsocs-Team/driver.sdk_go/pkg/objects"
)

func main() {
    log.Println("Starting temperature sensor driver...")

    c, err := client.New()
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    log.Println("Client initialized successfully")

    sensor := createTemperatureSensor()

    err = c.RegisterObject(sensor)
    if err != nil {
        log.Fatalf("Failed to register sensor: %v", err)
    }
    log.Println("Sensor registered successfully")

    go simulateTemperatureUpdates(sensor)

    log.Println("Driver ready, listening for requests...")
    err = c.ListenConfig()
    if err != nil {
        log.Fatalf("ListenConfig error: %v", err)
    }
}

func createTemperatureSensor() objects.SensorObject {
    params := objects.NewSensorObjectParams{
        Metadata: objects.ObjectMetadata{
            ObjectID: "temp_sensor_01",
            Name:     "Living Room Temperature",
            Domain:   "temperature",
            DeviceID: "demo_device_001",
            Tags:     []string{"temperature", "indoor"},
        },
        SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
            log.Println("Sensor setup called")
            sensor := obj.(objects.SensorObject)
            sensor.SetSensorType(objects.SensorObjectTypeNumber)
            sensor.SetUnitOfMeasurement("°C")
            sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)
            sensor.SetValue("20.0")
            log.Println("Sensor setup completed")
            return nil
        },
    }
    return objects.NewSensorObject(params)
}

func simulateTemperatureUpdates(sensor objects.SensorObject) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    temperature := 20.0

    for range ticker.C {
        temperature += (float64(time.Now().Unix()%3) - 1) * 0.5
        err := sensor.SetValue(fmt.Sprintf("%.1f", temperature))
        if err != nil {
            log.Printf("Error updating temperature: %v", err)
        } else {
            log.Printf("Temperature updated: %.1f°C", temperature)
        }
    }
}
```

</details>
