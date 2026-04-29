# Your First Driver

In this tutorial, you'll build a complete, working driver that creates a temperature sensor and integrates with the Netsocs platform. This hands-on guide will teach you the fundamental concepts while building something real.

## What You'll Build

A production-ready driver that:
- ✅ Connects to the Netsocs platform
- ✅ Creates a temperature sensor object
- ✅ Updates sensor values in real-time
- ✅ Handles platform configuration requests
- ✅ Implements proper error handling
- ✅ Follows SDK best practices

**⏱️ Time to complete**: 15-20 minutes

## Prerequisites

- ✅ Completed the [Installation Guide](installation.md)
- ✅ Have your `driver.netsocs.json` configuration file ready
- ✅ Basic understanding of Go programming

## Step 1: Project Setup

If you haven't already, create a new driver project:

```bash
# Create project directory
mkdir temperature-driver
cd temperature-driver

# Initialize Go module
go mod init github.com/myorg/temperature-driver

# Install the SDK
go get github.com/Netsocs-Team/driver.sdk_go
```

Create the basic project structure:

```bash
# Create directories
mkdir -p config devices objects

# Create main files
touch main.go
touch config/handlers.go
touch devices/simulator.go
touch objects/temperature_sensor.go
```

## Step 2: Configuration File

Create `driver.netsocs.json` with your credentials:

```json
{
  "driver_key": "YOUR_DRIVER_KEY_HERE",
  "driver_hub_host": "https://your-platform.netsocs.com/api/netsocs/dh",
  "token": "YOUR_AUTH_TOKEN",
  "driver_id": "YOUR_DRIVER_ID", 
  "site_id": "YOUR_SITE_ID",
  "name": "Temperature Sensor Driver",
  "version": "1.0.0",
  "driver_binary_filename": "temperature-driver",
  "documentation_url": "https://github.com/myorg/temperature-driver",
  "settings_available": [
    "actionPingDevice",
    "requestCreateObjects"
  ],
  "log_level": "info",
  "device_models_supported_all": true,
  "device_firmwares_supported_all": true
}
```

⚠️ **Security Note**: Replace the placeholder values with your actual credentials. Never commit this file to version control.

## Step 3: Device Simulator

First, let's create a simple device simulator that generates temperature readings.

**File**: `devices/simulator.go`

```go
package devices

import (
    "fmt"
    "math"
    "math/rand"
    "sync"
    "time"
)

// TemperatureDevice simulates a temperature sensor device
type TemperatureDevice struct {
    ID          string
    Name        string
    Location    string
    BaseTemp    float64
    isRunning   bool
    currentTemp float64
    mu          sync.RWMutex
    stopChan    chan struct{}
}

// NewTemperatureDevice creates a new temperature device simulator
func NewTemperatureDevice(id, name, location string, baseTemp float64) *TemperatureDevice {
    return &TemperatureDevice{
        ID:       id,
        Name:     name,
        Location: location,
        BaseTemp: baseTemp,
        stopChan: make(chan struct{}),
    }
}

// Start begins temperature simulation
func (td *TemperatureDevice) Start() {
    td.mu.Lock()
    if td.isRunning {
        td.mu.Unlock()
        return
    }
    td.isRunning = true
    td.currentTemp = td.BaseTemp
    td.mu.Unlock()

    go td.simulateTemperature()
}

// Stop ends temperature simulation
func (td *TemperatureDevice) Stop() {
    td.mu.Lock()
    defer td.mu.Unlock()
    
    if !td.isRunning {
        return
    }
    
    td.isRunning = false
    close(td.stopChan)
}

// GetTemperature returns the current temperature
func (td *TemperatureDevice) GetTemperature() float64 {
    td.mu.RLock()
    defer td.mu.RUnlock()
    return td.currentTemp
}

// IsOnline returns true if the device is running
func (td *TemperatureDevice) IsOnline() bool {
    td.mu.RLock()
    defer td.mu.RUnlock()
    return td.isRunning
}

// simulateTemperature generates realistic temperature variations
func (td *TemperatureDevice) simulateTemperature() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    startTime := time.Now()
    
    for {
        select {
        case <-td.stopChan:
            return
        case <-ticker.C:
            td.updateTemperature(startTime)
        }
    }
}

// updateTemperature calculates a new temperature value with realistic variations
func (td *TemperatureDevice) updateTemperature(startTime time.Time) {
    td.mu.Lock()
    defer td.mu.Unlock()
    
    // Simulate daily temperature cycle (24-hour period)
    elapsed := time.Since(startTime)
    hourOfDay := (elapsed.Hours()) // Accelerated for demo
    
    // Daily temperature variation (sine wave)
    dailyVariation := 3.0 * math.Sin((hourOfDay/24.0)*2*math.PI-math.Pi/2)
    
    // Random noise (-0.5 to +0.5 degrees)
    noise := (rand.Float64() - 0.5)
    
    // Calculate new temperature
    td.currentTemp = td.BaseTemp + dailyVariation + noise
    
    // Ensure reasonable bounds
    if td.currentTemp < -20 {
        td.currentTemp = -20
    } else if td.currentTemp > 50 {
        td.currentTemp = 50
    }
}

// DeviceManager manages multiple temperature devices
type DeviceManager struct {
    devices map[string]*TemperatureDevice
    mu      sync.RWMutex
}

// NewDeviceManager creates a new device manager
func NewDeviceManager() *DeviceManager {
    return &DeviceManager{
        devices: make(map[string]*TemperatureDevice),
    }
}

// AddDevice adds a temperature device to the manager
func (dm *DeviceManager) AddDevice(device *TemperatureDevice) {
    dm.mu.Lock()
    defer dm.mu.Unlock()
    dm.devices[device.ID] = device
}

// GetDevice retrieves a device by ID
func (dm *DeviceManager) GetDevice(id string) (*TemperatureDevice, bool) {
    dm.mu.RLock()
    defer dm.mu.RUnlock()
    device, exists := dm.devices[id]
    return device, exists
}

// GetAllDevices returns all managed devices
func (dm *DeviceManager) GetAllDevices() []*TemperatureDevice {
    dm.mu.RLock()
    defer dm.mu.RUnlock()
    
    devices := make([]*TemperatureDevice, 0, len(dm.devices))
    for _, device := range dm.devices {
        devices = append(devices, device)
    }
    return devices
}

// StartAll starts all devices
func (dm *DeviceManager) StartAll() {
    dm.mu.RLock()
    defer dm.mu.RUnlock()
    
    for _, device := range dm.devices {
        device.Start()
    }
}

// StopAll stops all devices
func (dm *DeviceManager) StopAll() {
    dm.mu.RLock()
    defer dm.mu.RUnlock()
    
    for _, device := range dm.devices {
        device.Stop()
    }
}

// Ping tests device connectivity (always succeeds for simulator)
func (dm *DeviceManager) Ping(deviceID string) error {
    device, exists := dm.GetDevice(deviceID)
    if !exists {
        return fmt.Errorf("device %s not found", deviceID)
    }
    
    if !device.IsOnline() {
        return fmt.Errorf("device %s is offline", deviceID)
    }
    
    return nil
}
```

## Step 4: Temperature Sensor Object

Now let's create a temperature sensor object that uses our device simulator.

**File**: `objects/temperature_sensor.go`

```go
package objects

import (
    "fmt"
    "log"
    "strconv"
    "time"

    "github.com/Netsocs-Team/driver.sdk_go/pkg/objects"
    
    "github.com/myorg/temperature-driver/devices"
)

// TemperatureSensor wraps a Netsocs SensorObject with temperature-specific functionality
type TemperatureSensor struct {
    sensorObject objects.SensorObject
    device       *devices.TemperatureDevice
    deviceMgr    *devices.DeviceManager
    stopChan     chan struct{}
}

// NewTemperatureSensor creates a new temperature sensor object
func NewTemperatureSensor(deviceID, objectID, name, location string, deviceMgr *devices.DeviceManager) *TemperatureSensor {
    // Create the underlying device simulator
    device := devices.NewTemperatureDevice(deviceID, name, location, 20.0) // 20°C base temperature
    deviceMgr.AddDevice(device)
    
    // Create the Netsocs sensor object
    sensorObject := objects.NewSensorObject(objects.NewSensorObjectParams{
        Metadata: objects.ObjectMetadata{
            ObjectID: objectID,
            Name:     name,
            Domain:   "temperature",
            DeviceID: deviceID,
            Tags:     []string{"temperature", "indoor", location},
        },
        SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
            log.Printf("Setting up temperature sensor: %s", name)
            
            sensor := obj.(objects.SensorObject)
            
            // Configure sensor properties
            if err := sensor.SetSensorType(objects.SensorObjectTypeNumber); err != nil {
                return fmt.Errorf("failed to set sensor type: %w", err)
            }
            
            if err := sensor.SetUnitOfMeasurement("°C"); err != nil {
                return fmt.Errorf("failed to set unit of measurement: %w", err)
            }
            
            // Set initial state
            if err := sensor.SetState(objects.SENSOR_STATE_MEASUREMENT); err != nil {
                return fmt.Errorf("failed to set initial state: %w", err)
            }
            
            // Set initial temperature
            if err := sensor.SetValue("20.0"); err != nil {
                return fmt.Errorf("failed to set initial value: %w", err)
            }
            
            log.Printf("Temperature sensor %s setup completed", name)
            return nil
        },
    })
    
    return &TemperatureSensor{
        sensorObject: sensorObject,
        device:       device,
        deviceMgr:    deviceMgr,
        stopChan:     make(chan struct{}),
    }
}

// GetSensorObject returns the underlying Netsocs sensor object
func (ts *TemperatureSensor) GetSensorObject() objects.SensorObject {
    return ts.sensorObject
}

// Start begins temperature monitoring and updates
func (ts *TemperatureSensor) Start() {
    log.Printf("Starting temperature sensor: %s", ts.sensorObject.GetMetadata().Name)
    
    // Start the device simulator
    ts.device.Start()
    
    // Start the update loop
    go ts.updateLoop()
}

// Stop ends temperature monitoring
func (ts *TemperatureSensor) Stop() {
    log.Printf("Stopping temperature sensor: %s", ts.sensorObject.GetMetadata().Name)
    
    // Stop the update loop
    close(ts.stopChan)
    
    // Stop the device simulator
    ts.device.Stop()
}

// updateLoop continuously updates the sensor value from the device
func (ts *TemperatureSensor) updateLoop() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ts.stopChan:
            return
        case <-ticker.C:
            ts.updateTemperature()
        }
    }
}

// updateTemperature reads from device and updates the sensor object
func (ts *TemperatureSensor) updateTemperature() {
    // Get current temperature from device
    temperature := ts.device.GetTemperature()
    
    // Format temperature to 1 decimal place
    tempStr := strconv.FormatFloat(temperature, 'f', 1, 64)
    
    // Update sensor value
    if err := ts.sensorObject.SetValue(tempStr); err != nil {
        log.Printf("Error updating temperature for %s: %v", 
            ts.sensorObject.GetMetadata().Name, err)
        return
    }
    
    // Update additional state attributes
    attributes := map[string]string{
        "temperature":        tempStr,
        "unit_of_measurement": "°C",
        "device_status":      "online",
        "last_updated":       time.Now().Format(time.RFC3339),
        "location":           ts.device.Location,
    }
    
    if err := ts.sensorObject.UpdateStateAttributes(attributes); err != nil {
        log.Printf("Error updating attributes for %s: %v", 
            ts.sensorObject.GetMetadata().Name, err)
        return
    }
    
    log.Printf("Temperature updated: %s = %s°C", 
        ts.sensorObject.GetMetadata().Name, tempStr)
}

// GetCurrentTemperature returns the current temperature value
func (ts *TemperatureSensor) GetCurrentTemperature() float64 {
    return ts.device.GetTemperature()
}

// IsOnline returns true if the sensor device is online
func (ts *TemperatureSensor) IsOnline() bool {
    return ts.device.IsOnline()
}
```

## Step 5: Configuration Handlers

Create handlers for platform requests.

**File**: `config/handlers.go`

```go
package config

import (
    "fmt"
    "log"
    "strconv"

    "github.com/Netsocs-Team/driver.sdk_go/pkg/client"
    "github.com/Netsocs-Team/driver.sdk_go/pkg/config"
    
    "github.com/myorg/temperature-driver/devices"
    "github.com/myorg/temperature-driver/objects"
)

// RegisterHandlers registers all configuration handlers with the client
func RegisterHandlers(c *client.NetsocsDriverClient, deviceMgr *devices.DeviceManager) error {
    // Map of handler keys to handler functions
    handlers := map[config.NetsocsConfigKey]config.FuncConfigHandler{
        config.ACTION_PING_DEVICE:    handlePingDevice(deviceMgr),
        config.REQUEST_CREATE_OBJECTS: handleCreateObjects(c, deviceMgr),
    }
    
    // Register each handler
    for key, handler := range handlers {
        if err := c.AddConfigHandler(key, handler); err != nil {
            return fmt.Errorf("failed to register handler %s: %w", key, err)
        }
        log.Printf("Registered handler: %s", key)
    }
    
    return nil
}

// handlePingDevice handles device ping requests
func handlePingDevice(deviceMgr *devices.DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        log.Printf("Handling ACTION_PING_DEVICE for device ID: %d", msg.DeviceData.ID)
        
        deviceID := strconv.Itoa(msg.DeviceData.ID)
        
        // Test device connectivity
        err := deviceMgr.Ping(deviceID)
        if err != nil {
            log.Printf("Ping failed for device %s: %v", deviceID, err)
            return map[string]interface{}{
                "status": false,
                "error":  true,
                "msg":    fmt.Sprintf("Device unreachable: %v", err),
            }, nil
        }
        
        log.Printf("Ping successful for device %s", deviceID)
        return map[string]interface{}{
            "status": true,
            "error":  false,
            "msg":    "Device is online and responding",
        }, nil
    }
}

// handleCreateObjects handles object creation requests
func handleCreateObjects(c *client.NetsocsDriverClient, deviceMgr *devices.DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        log.Printf("Handling REQUEST_CREATE_OBJECTS for device: %s (ID: %d)", 
            msg.DeviceData.Name, msg.DeviceData.ID)
        
        deviceID := strconv.Itoa(msg.DeviceData.ID)
        deviceName := msg.DeviceData.Name
        
        // Create temperature sensors for different locations
        sensors := []*objects.TemperatureSensor{
            objects.NewTemperatureSensor(
                deviceID,
                fmt.Sprintf("temp_sensor_%s_living_room", deviceID),
                fmt.Sprintf("%s - Living Room", deviceName),
                "living_room",
                deviceMgr,
            ),
            objects.NewTemperatureSensor(
                deviceID,
                fmt.Sprintf("temp_sensor_%s_bedroom", deviceID),
                fmt.Sprintf("%s - Bedroom", deviceName),
                "bedroom",
                deviceMgr,
            ),
            objects.NewTemperatureSensor(
                deviceID,
                fmt.Sprintf("temp_sensor_%s_kitchen", deviceID),
                fmt.Sprintf("%s - Kitchen", deviceName),
                "kitchen",
                deviceMgr,
            ),
        }
        
        // Register each sensor with the platform
        for _, sensor := range sensors {
            if err := c.RegisterObject(sensor.GetSensorObject()); err != nil {
                log.Printf("Failed to register sensor %s: %v", 
                    sensor.GetSensorObject().GetMetadata().Name, err)
                return nil, fmt.Errorf("failed to register sensor: %w", err)
            }
            
            // Start the sensor
            sensor.Start()
            
            log.Printf("Successfully registered and started sensor: %s", 
                sensor.GetSensorObject().GetMetadata().Name)
        }
        
        // Set device state to online
        if err := c.SetDeviceState(msg.DeviceData.ID, client.DeviceStateOnline); err != nil {
            log.Printf("Failed to set device state: %v", err)
        }
        
        log.Printf("Successfully created %d temperature sensors for device %s", 
            len(sensors), deviceName)
        
        return map[string]interface{}{
            "success":      true,
            "sensors_created": len(sensors),
            "message":      fmt.Sprintf("Created %d temperature sensors", len(sensors)),
        }, nil
    }
}
```

## Step 6: Main Driver Application

Now let's tie everything together in the main application.

**File**: `main.go`

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/Netsocs-Team/driver.sdk_go/pkg/client"
    
    "github.com/myorg/temperature-driver/config"
    "github.com/myorg/temperature-driver/devices"
)

var (
    version = "1.0.0"
    author  = "Your Organization"
    description = "Temperature Sensor Driver for Netsocs"
)

func main() {
    log.Printf("===========================================")
    log.Printf(" Starting %s", description)
    log.Printf(" Version: %s", version)
    log.Printf(" Author: %s", author)
    log.Printf("===========================================")
    
    // Initialize device manager
    deviceMgr := devices.NewDeviceManager()
    
    // Initialize SDK client
    log.Println("Initializing Netsocs SDK client...")
    sdkClient, err := client.New()
    if err != nil {
        log.Fatalf("Failed to create SDK client: %v", err)
    }
    log.Println("✓ SDK client initialized successfully")
    
    // Set driver metadata
    sdkClient.SetDriverVersion(version)
    sdkClient.SetDriverDocumentation("https://github.com/myorg/temperature-driver")
    
    // Register configuration handlers
    log.Println("Registering configuration handlers...")
    if err := config.RegisterHandlers(sdkClient, deviceMgr); err != nil {
        log.Fatalf("Failed to register handlers: %v", err)
    }
    log.Println("✓ Configuration handlers registered")
    
    // Setup graceful shutdown
    setupGracefulShutdown(deviceMgr)
    
    // Start listening for platform requests
    log.Printf("===========================================")
    log.Printf(" Driver ready, listening for requests...")
    log.Printf("===========================================")
    
    // This is a blocking call that keeps the driver running
    if err := sdkClient.ListenConfig(); err != nil {
        log.Fatalf("ListenConfig error: %v", err)
    }
}

// setupGracefulShutdown handles cleanup when the driver is terminated
func setupGracefulShutdown(deviceMgr *devices.DeviceManager) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        log.Println("\n===========================================")
        log.Println(" Shutting down driver...")
        log.Println("===========================================")
        
        // Stop all devices
        log.Println("Stopping all devices...")
        deviceMgr.StopAll()
        log.Println("✓ All devices stopped")
        
        log.Println("Driver shutdown complete")
        os.Exit(0)
    }()
}
```

## Step 7: Build and Test

### Download Dependencies

```bash
go mod tidy
```

### Build the Driver

```bash
go build -o temperature-driver
```

### Run the Driver

```bash
./temperature-driver
```

Expected output:
```
===========================================
 Starting Temperature Sensor Driver for Netsocs
 Version: 1.0.0
 Author: Your Organization
===========================================
Initializing Netsocs SDK client...
✓ SDK client initialized successfully
Registering configuration handlers...
Registered handler: ACTION_PING_DEVICE
Registered handler: REQUEST_CREATE_OBJECTS
✓ Configuration handlers registered
===========================================
 Driver ready, listening for requests...
===========================================
```

## Step 8: Test in the Platform

### 1. Add Device in Platform UI

1. **Log in to the Netsocs platform**
2. **Navigate to Devices → Add Device**
3. **Select your driver** from the dropdown
4. **Fill in device details:**
   - Name: "Temperature Monitor"
   - IP: "127.0.0.1" (or any IP for simulator)
   - Port: 80
   - Username: "admin"
   - Password: "password"

### 2. Test Device Connection

1. **Click "Test Connection"** - should trigger `ACTION_PING_DEVICE`
2. **Verify success message** appears

### 3. Create Objects

1. **Click "Create Objects"** - triggers `REQUEST_CREATE_OBJECTS`
2. **Wait for completion** - should see success message
3. **Navigate to Objects** - should see 3 temperature sensors:
   - Temperature Monitor - Living Room
   - Temperature Monitor - Bedroom  
   - Temperature Monitor - Kitchen

### 4. Monitor Real-Time Updates

1. **Open any sensor** in the platform
2. **Watch the temperature values** update every 5 seconds
3. **Check the state attributes** for additional information

## Step 9: Understanding What Happens

### Driver Lifecycle

```
Platform Request → Handler → Device Interaction → Object Update → Platform Response
```

1. **Platform sends request** via WebSocket to your driver
2. **Handler processes request** (ping device, create objects)
3. **Driver interacts with devices** (real or simulated)
4. **Objects are created/updated** with current states
5. **Response sent back** to platform

### Object Updates

```
Device Simulator → Temperature Reading → Sensor Object → Platform State Update
```

1. **Device simulator** generates realistic temperature variations
2. **Update loop** reads temperature every 5 seconds
3. **Sensor object** updates value and attributes
4. **Platform receives** state updates via HTTP API

### Key SDK Concepts Demonstrated

- ✅ **Client initialization** and configuration
- ✅ **Configuration handlers** for platform requests
- ✅ **Object creation** and registration
- ✅ **State management** with values and attributes
- ✅ **Background tasks** for continuous updates
- ✅ **Error handling** and logging
- ✅ **Graceful shutdown** and cleanup

## Next Steps

Congratulations! You've built a complete, working Netsocs driver. Here's what to explore next:

### Immediate Next Steps

1. **[Understanding Objects](objects.md)** - Deep dive into object types and lifecycle
2. **[Configuration Handlers](handlers.md)** - Learn about all 70+ available handlers
3. **[Integration Guides](integrations/)** - Adapt this pattern for real devices

### Extend Your Driver

Try these enhancements to deepen your understanding:

#### Add More Sensor Types

```go
// Add humidity sensor
humiditySensor := objects.NewSensorObject(objects.NewSensorObjectParams{
    Metadata: objects.ObjectMetadata{
        ObjectID: "humidity_01",
        Name:     "Living Room Humidity",
        Domain:   "humidity",
    },
    SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
        sensor := obj.(objects.SensorObject)
        sensor.SetSensorType(objects.SensorObjectTypeNumber)
        sensor.SetUnitOfMeasurement("%")
        return nil
    },
})
```

#### Add Switch Control

```go
// Add controllable heater
heater := objects.NewSwitchObject(objects.NewSwitchObjectParams{
    Metadata: objects.ObjectMetadata{
        ObjectID: "heater_01",
        Name:     "Living Room Heater",
        Domain:   "climate",
    },
    TurnOnMethod: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
        log.Println("Turning heater ON")
        return nil
    },
    TurnOffMethod: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
        log.Println("Turning heater OFF")
        return nil
    },
})
```

#### Add Events

```go
// Register temperature alert event
alertEvent := objects.EventType{
    Domain:             "temperature",
    EventType:          "temperature_alert",
    DisplayName:        "Temperature Alert",
    DisplayDescription: "Temperature exceeded threshold",
    EventLevel:         "warning",
    Color:              "#FFA500",
    ShowColor:          true,
    Origin:             "driver",
}

sdkClient.AddEventTypes([]objects.EventType{alertEvent})

// Dispatch event when temperature is too high
if temperature > 30.0 {
    eventData := objects.Event{
        ObjectIDs: []string{sensorObjectID},
        Properties: map[string]string{
            "temperature": fmt.Sprintf("%.1f", temperature),
            "threshold":   "30.0",
            "location":    location,
        },
    }
    sdkClient.DispatchEvent("temperature", "temperature_alert", eventData)
}
```

### Real Device Integration

Replace the simulator with real device communication:

```go
// Example: HTTP-based temperature device
func (td *TemperatureDevice) GetTemperature() (float64, error) {
    resp, err := http.Get(fmt.Sprintf("http://%s/api/temperature", td.IP))
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()
    
    var data struct {
        Temperature float64 `json:"temperature"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return 0, err
    }
    
    return data.Temperature, nil
}
```

## Troubleshooting

### Common Issues

#### "Failed to create client"
- ✅ Verify `driver.netsocs.json` exists in working directory
- ✅ Check JSON syntax with `cat driver.netsocs.json | python -m json.tool`
- ✅ Ensure all required fields are present

#### "Failed to register object"
- ✅ Check ObjectID is unique within your driver
- ✅ Verify network connectivity to platform
- ✅ Check driver credentials are correct

#### Objects not appearing in platform
- ✅ Verify `REQUEST_CREATE_OBJECTS` handler completed successfully
- ✅ Check driver logs for registration errors
- ✅ Ensure driver is activated in platform

#### Temperature not updating
- ✅ Check device simulator is started
- ✅ Verify update loop is running (check logs)
- ✅ Ensure sensor object SetValue calls succeed

### Debug Mode

Enable debug logging by updating your configuration:

```json
{
  "log_level": "debug"
}
```

Or add debug logging to your code:

```go
log.SetLevel(log.DebugLevel)
log.Debug("Temperature update:", temperature)
```

## Summary

You've successfully built a complete Netsocs driver that demonstrates:

- **SDK Integration**: Proper client initialization and configuration
- **Object Management**: Creating and managing sensor objects
- **State Updates**: Real-time value and attribute updates  
- **Handler Implementation**: Responding to platform requests
- **Background Processing**: Continuous device monitoring
- **Error Handling**: Robust error management and logging

This foundation prepares you to build drivers for any IoT device or system. The patterns you've learned here scale to complex integrations with cameras, access control systems, alarm panels, and cloud services.

**Ready for more?** Continue with [Understanding Objects](objects.md) to master the object system, or jump to [Integration Guides](integrations/) to see real-world examples for specific device types.