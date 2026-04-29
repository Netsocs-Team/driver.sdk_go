# Understanding Objects

Objects are the fundamental building blocks of the Netsocs Driver SDK. They represent controllable or monitorable entities in your integrated systems. This comprehensive guide explains what objects are, how they work, and how to use them effectively.

## What is an Object?

An **object** represents a real-world entity that can be monitored, controlled, or both. Objects bridge the gap between physical devices and the Netsocs platform's digital representation.

### Examples of Objects

| Object Type | Real-World Entity | Platform Representation |
|-------------|-------------------|-------------------------|
| **SensorObject** | Temperature sensor | Current temperature reading |
| **VideoChannelObject** | IP camera | Live stream, snapshots, recordings |
| **LockObject** | Electronic door lock | Lock/unlock status and control |
| **SwitchObject** | Light switch or relay | On/off status and control |
| **AlarmPanelObject** | Security panel | Armed/disarmed status |
| **ReaderObject** | Card/biometric reader | Access events and credential management |

## Object Architecture

Every object in the Netsocs SDK implements the `RegistrableObject` interface:

```go
type RegistrableObject interface {
    Setup(ObjectController) error
    GetAvailableStates() []string
    GetAvailableActions() []ObjectAction
    RunAction(id, action string, payload []byte) (map[string]string, error)
    GetMetadata() ObjectMetadata
    SetState(state string) error
    UpdateStateAttributes(attributes map[string]string) error
}
```

### Core Components

#### 1. Metadata
Identifies and classifies the object:

```go
type ObjectMetadata struct {
    ObjectID string            // Unique identifier within driver
    Name     string            // Human-readable display name
    Type     string            // Object type (sensor, switch, camera, etc.)
    Domain   string            // Logical grouping (temperature, security, etc.)
    DeviceID string            // Parent device identifier
    Tags     []string          // Searchable keywords
    ParentID string            // Parent object (for hierarchies)
    I18n     map[string]string // Internationalization support
}
```

#### 2. States
The current condition or status:

```go
// Examples of states
sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)     // "sensor.state.measurement"
switch.SetState(objects.SWITCH_STATE_ON)              // "switch.state.on"
lock.SetState(objects.LOCK_STATE_LOCKED)              // "lock.state.locked"
camera.SetState(objects.VIDEO_CHANNEL_STATE_STREAMING) // "video_channel.state.streaming"
```

#### 3. State Attributes
Additional key-value properties:

```go
sensor.UpdateStateAttributes(map[string]string{
    "value":              "23.5",
    "unit_of_measurement": "°C",
    "battery_level":      "85%",
    "signal_strength":    "strong",
    "last_updated":       time.Now().Format(time.RFC3339),
})
```

#### 4. Actions
Operations that can be performed:

```go
// Switch actions
{Action: "switch.action.turn_on", Domain: "switch"}
{Action: "switch.action.turn_off", Domain: "switch"}
{Action: "switch.action.toggle", Domain: "switch"}

// Camera actions
{Action: "video_channel.action.snapshot", Domain: "camera"}
{Action: "video_channel.action.ptz_control", Domain: "camera"}
```

## Object Lifecycle

Understanding the object lifecycle helps you implement objects correctly:

```
┌─────────────────┐
│  1. Creation    │  objects.NewSensorObject(params)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 2. Registration │  client.RegisterObject(obj)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   3. Setup()    │  Initialize, set defaults, start tasks
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 4. Active Phase │  Ready for actions, state updates
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  5. Operation Phase                     │
│  • RunAction() - User/automation        │
│  • SetState() - Condition changes       │
│  • UpdateStateAttributes() - Data       │
└─────────────────────────────────────────┘
```

### Phase 1: Creation

Objects are created using constructor functions provided by the SDK:

```go
// Create a temperature sensor
sensor := objects.NewSensorObject(objects.NewSensorObjectParams{
    Metadata: objects.ObjectMetadata{
        ObjectID: "temp_sensor_01",
        Name:     "Living Room Temperature",
        Domain:   "temperature",
        DeviceID: "device_123",
    },
    SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
        // Initialization logic here
        return nil
    },
})
```

**At this point**: Object exists in memory but the platform doesn't know about it.

### Phase 2: Registration

Objects are registered with the platform using the SDK client:

```go
err := client.RegisterObject(sensor)
if err != nil {
    return fmt.Errorf("failed to register sensor: %w", err)
}
```

**What happens**:
1. SDK sends object metadata to the platform
2. Platform creates a database entry
3. Object becomes visible in the UI
4. Platform prepares to receive state updates

### Phase 3: Setup

The SDK automatically calls the object's `Setup()` method (or `SetupFn` for built-in objects):

```go
SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
    sensor := obj.(objects.SensorObject)
    
    // Configure sensor properties
    sensor.SetSensorType(objects.SensorObjectTypeNumber)
    sensor.SetUnitOfMeasurement("°C")
    
    // Set initial state
    sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)
    sensor.SetValue("20.0")
    
    // Start background tasks
    go pollDeviceAndUpdate(sensor)
    
    return nil
}
```

**Purpose**: Initialize the object, set default values, and start any background processes.

### Phase 4: Active Phase

The object is now registered and ready to:
- Receive actions from users or automations
- Update its state based on device changes
- Provide real-time data to the platform

### Phase 5: Operation Phase

The object operates continuously:

```go
// State updates (from device changes)
sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)
sensor.SetValue("24.3")

// Attribute updates (additional context)
sensor.UpdateStateAttributes(map[string]string{
    "battery_level": "78%",
    "signal":        "good",
})

// Action handling (from user/automation)
func (s *switchObject) RunAction(id, action string, payload []byte) (map[string]string, error) {
    switch action {
    case "switch.action.turn_on":
        return s.turnOn()
    case "switch.action.turn_off":
        return s.turnOff()
    }
}
```

## Built-in Object Types

The SDK provides 25+ built-in object types. Here are the most commonly used:

### SensorObject

For measurements, readings, and monitoring.

```go
sensor := objects.NewSensorObject(objects.NewSensorObjectParams{
    Metadata: objects.ObjectMetadata{
        ObjectID: "temp_01",
        Name:     "Temperature Sensor",
        Domain:   "temperature",
        DeviceID: "device_123",
    },
    SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
        sensor := obj.(objects.SensorObject)
        
        // Configure sensor type
        sensor.SetSensorType(objects.SensorObjectTypeNumber) // Number, Text, Binary, Battery
        sensor.SetUnitOfMeasurement("°C")
        
        // Set initial state and value
        sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)
        sensor.SetValue("20.0")
        
        return nil
    },
})
```

**Use cases**: Temperature, humidity, motion, smoke, water leak, light level, air quality, pressure, etc.

**Key methods**:
- `SetValue(value string)` - Update the reading
- `SetSensorType(type)` - Number, Text, Binary, Battery
- `SetUnitOfMeasurement(unit)` - °C, %, lux, ppm, etc.
- `Increment()` / `Decrement()` - For counters

**Available states**:
- `SENSOR_STATE_MEASUREMENT` - Current reading
- `SENSOR_STATE_TOTAL` - Cumulative value
- `SENSOR_STATE_TOTAL_INCREASING` - Ever-increasing counter

### SwitchObject

For controllable on/off devices.

```go
sw := objects.NewSwitchObject(objects.NewSwitchObjectParams{
    Metadata: objects.ObjectMetadata{
        ObjectID: "relay_01",
        Name:     "Living Room Light",
        Domain:   "light",
        DeviceID: "device_123",
    },
    TurnOnMethod: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
        // Send command to physical device
        return sendDeviceCommand("RELAY_1_ON")
    },
    TurnOffMethod: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
        // Send command to physical device
        return sendDeviceCommand("RELAY_1_OFF")
    },
})
```

**Use cases**: Lights, relays, sirens, garage doors, fans, heaters, etc.

**Available actions**:
- `SWITCH_ACTION_TURN_ON` - Turn device on
- `SWITCH_ACTION_TURN_OFF` - Turn device off
- `SWITCH_ACTION_TOGGLE` - Toggle current state

**Available states**:
- `SWITCH_STATE_ON` - Device is on
- `SWITCH_STATE_OFF` - Device is off
- `SWITCH_STATE_UNKNOWN` - State is unclear

### VideoChannelObject

For cameras, video streams, and recording systems.

```go
camera := objects.NewVideoChannelObject(objects.NewVideoChannelObjectProps{
    Metadata: objects.ObjectMetadata{
        ObjectID: "camera_01_ch1",
        Name:     "Front Entrance Camera",
        Domain:   "camera",
        DeviceID: "nvr_001",
    },
    StreamID:    "rtsp://192.168.1.10:554/stream1",
    VideoEngine: "video_engine_01",
    PTZ:         true, // Pan-Tilt-Zoom support
    SnapshotFn: func(vc objects.VideoChannelObject, oc objects.ObjectController,
                     payload objects.SnapshotActionPayload) (string, error) {
        // Capture snapshot from camera
        imageData, err := captureSnapshot(vc.GetMetadata().ObjectID)
        if err != nil {
            return "", err
        }
        
        // Upload image and return URL
        imageURL, err := uploadImage(imageData)
        return imageURL, err
    },
})
```

**Use cases**: IP cameras, NVRs, DVRs, webcams, security cameras

**Available actions**:
- `VIDEO_CHANNEL_ACTION_SNAPSHOT` - Capture still image
- `VIDEO_CHANNEL_ACTION_VIDEOCLIP` - Record video clip
- `VIDEO_CHANNEL_ACTION_PTZ_CONTROL` - Control camera movement

**Available states**:
- `VIDEO_CHANNEL_STATE_STREAMING` - Actively streaming
- `VIDEO_CHANNEL_STATE_IDLE` - Not streaming
- `VIDEO_CHANNEL_STATE_RECORDING` - Recording video

### LockObject

For access control and security locks.

```go
lock := objects.NewLockObject(objects.NewLockObjectParams{
    Metadata: objects.ObjectMetadata{
        ObjectID: "door_lock_01",
        Name:     "Main Entrance Lock",
        Domain:   "access_control",
        DeviceID: "device_123",
    },
    LockMethod: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
        return sendLockCommand("LOCK")
    },
    UnlockMethod: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
        return sendLockCommand("UNLOCK")
    },
})
```

**Use cases**: Electronic locks, magnetic locks, electric strikes, smart locks

**Available actions**:
- `LOCK_ACTION_LOCK` - Secure the lock
- `LOCK_ACTION_UNLOCK` - Release the lock

**Available states**:
- `LOCK_STATE_LOCKED` - Lock is secured
- `LOCK_STATE_UNLOCKED` - Lock is released
- `LOCK_STATE_UNKNOWN` - State is unclear

### AlarmPanelObject

For security and alarm systems.

```go
panel := objects.NewAlarmPanelObject(objects.NewAlarmPanelObjectProps{
    Metadata: objects.ObjectMetadata{
        ObjectID: "alarm_panel_01",
        Name:     "Main Security Panel",
        Domain:   "alarm",
        DeviceID: "panel_001",
    },
})
```

**Use cases**: Security panels, fire alarm systems, intrusion detection

**Available states**:
- `ALARM_PANEL_STATE_DISARMED` - System is disarmed
- `ALARM_PANEL_STATE_ARMED_HOME` - Armed in home mode
- `ALARM_PANEL_STATE_ARMED_AWAY` - Armed in away mode
- `ALARM_PANEL_STATE_TRIGGERED` - Alarm is active

### ReaderObject

For credential readers and access control.

```go
reader := objects.NewReaderObject(objects.NewReaderObjectParams{
    Metadata: objects.ObjectMetadata{
        ObjectID: "reader_main_entrance",
        Name:     "Main Entrance Reader",
        Domain:   "access_control",
        DeviceID: "ac_panel_001",
    },
    SupportedCredentialTypes: []string{"card", "face", "fingerprint", "qr"},
})
```

**Use cases**: Card readers, biometric scanners, QR code readers, PIN pads

**Supported credential types**:
- `"card"` - RFID/proximity cards
- `"face"` - Facial recognition
- `"fingerprint"` - Fingerprint scanning
- `"qr"` - QR code scanning
- `"pin"` - PIN entry

## Object Best Practices

### Metadata Guidelines

#### ObjectID Design

Create stable, unique, descriptive IDs:

```go
// Good examples
ObjectID: "nvr_001_camera_ch1"      // NVR device, camera channel 1
ObjectID: "building_a_temp_lobby"   // Building A, temperature sensor in lobby
ObjectID: "door_main_entrance"      // Main entrance door
ObjectID: "reader_parking_gate"     // Parking gate card reader

// Avoid
ObjectID: "obj1"                    // Not descriptive
ObjectID: "sensor"                  // Not unique
ObjectID: "camera ch1"              // Contains spaces
ObjectID: "temp@lobby"              // Special characters
```

**Rules**:
- Use only alphanumeric characters, hyphens, and underscores
- Make IDs deterministic (same device = same ID across restarts)
- Include context (device, location, function)
- Keep under 64 characters

#### Domain Organization

Use consistent domains to group related objects:

```go
// Environmental monitoring
Domain: "temperature"
Domain: "humidity" 
Domain: "air_quality"

// Security systems
Domain: "camera"
Domain: "access_control"
Domain: "alarm"

// Building automation
Domain: "lighting"
Domain: "hvac"
Domain: "energy"
```

#### Effective Tagging

Use tags for filtering and organization:

```go
Tags: []string{"indoor", "temperature", "living_room", "critical"}
Tags: []string{"outdoor", "camera", "perimeter", "night_vision"}
Tags: []string{"access_control", "biometric", "high_security"}
```

### State Management

#### States vs State Attributes

**Use States for**:
- Primary operational condition
- Conditions that trigger automations
- Status changes users care about in history
- Enumerated values with specific meanings

**Use State Attributes for**:
- Measurement values
- Metadata (battery, signal, version)
- Frequently changing data
- Additional context information

```go
// Example: Temperature Sensor
// State: Type of measurement being taken
sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)

// Attributes: The actual data and metadata
sensor.UpdateStateAttributes(map[string]string{
    "value":              "23.5",        // The temperature reading
    "unit_of_measurement": "°C",          // Display unit
    "battery_level":      "85%",         // Device metadata
    "signal_strength":    "good",        // Connection quality
    "last_updated":       "2024-01-15T10:30:00Z", // Timestamp
    "calibration_offset": "0.2",         // Device configuration
})
```

#### State Update Patterns

**Immediate Updates**: For user actions and critical changes
```go
// User pressed switch - update immediately
switch.SetState(objects.SWITCH_STATE_ON)
```

**Batched Updates**: For frequent sensor data
```go
// Update every 30 seconds, not every reading
ticker := time.NewTicker(30 * time.Second)
for range ticker.C {
    sensor.SetValue(getCurrentReading())
}
```

**Event-Driven Updates**: For device notifications
```go
// Update when device sends notification
deviceClient.OnStateChange(func(newState string) {
    object.SetState(newState)
})
```

### Action Implementation

#### Robust Action Handlers

```go
func (s *switchObject) RunAction(id, action string, payload []byte) (map[string]string, error) {
    switch action {
    case "switch.action.turn_on":
        // 1. Validate current state
        if s.getCurrentState() == objects.SWITCH_STATE_ON {
            return map[string]string{
                "status": "already_on",
                "message": "Switch is already on",
            }, nil
        }
        
        // 2. Send command to device with timeout
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        
        err := s.deviceClient.TurnOn(ctx)
        if err != nil {
            return nil, fmt.Errorf("failed to turn on switch: %w", err)
        }
        
        // 3. Update state
        s.SetState(objects.SWITCH_STATE_ON)
        
        // 4. Return success response
        return map[string]string{
            "status": "success",
            "message": "Switch turned on successfully",
            "new_state": "on",
        }, nil
        
    case "switch.action.turn_off":
        // Similar implementation for turn off
        
    default:
        return nil, fmt.Errorf("unknown action: %s", action)
    }
}
```

#### Action Response Format

Always return structured responses:

```go
// Success response
return map[string]string{
    "status":    "success",
    "message":   "Operation completed",
    "new_state": "locked",
    "timestamp": time.Now().Format(time.RFC3339),
}, nil

// Error response (return error, not error in map)
return nil, fmt.Errorf("device unreachable: connection timeout")
```

### Background Tasks

#### Proper Goroutine Management

```go
type TemperatureSensor struct {
    sensor   objects.SensorObject
    device   DeviceClient
    stopChan chan struct{}
    wg       sync.WaitGroup
}

func (ts *TemperatureSensor) Start() {
    ts.stopChan = make(chan struct{})
    
    ts.wg.Add(1)
    go func() {
        defer ts.wg.Done()
        ts.updateLoop()
    }()
}

func (ts *TemperatureSensor) Stop() {
    close(ts.stopChan)
    ts.wg.Wait() // Wait for goroutine to finish
}

func (ts *TemperatureSensor) updateLoop() {
    ticker := time.NewTicker(30 * time.Second)
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
```

## Advanced Object Patterns

### Parent-Child Relationships

Create hierarchical object structures:

```go
// Parent: NVR device
nvr := objects.NewVideoEngineObject(objects.NewVideoEngineObjectProps{
    Metadata: objects.ObjectMetadata{
        ObjectID: "nvr_001",
        Name:     "Main NVR",
        Domain:   "video_system",
    },
})

// Children: Camera channels
for i := 1; i <= 16; i++ {
    camera := objects.NewVideoChannelObject(objects.NewVideoChannelObjectProps{
        Metadata: objects.ObjectMetadata{
            ObjectID: fmt.Sprintf("nvr_001_ch_%d", i),
            Name:     fmt.Sprintf("Camera Channel %d", i),
            Domain:   "camera",
            DeviceID: "nvr_001",
            ParentID: "nvr_001", // Link to parent
        },
        StreamID: fmt.Sprintf("rtsp://nvr.local:554/ch%d", i),
    })
    
    client.RegisterObject(camera)
}
```

### Dynamic Object Creation

Create objects based on device discovery:

```go
func (d *Driver) CreateObjectsFromDevice(deviceID string) error {
    // Discover device capabilities
    capabilities, err := d.deviceClient.GetCapabilities()
    if err != nil {
        return err
    }
    
    // Create objects based on what device supports
    for _, sensor := range capabilities.Sensors {
        sensorObj := objects.NewSensorObject(objects.NewSensorObjectParams{
            Metadata: objects.ObjectMetadata{
                ObjectID: fmt.Sprintf("%s_sensor_%s", deviceID, sensor.ID),
                Name:     sensor.Name,
                Domain:   sensor.Type,
                DeviceID: deviceID,
            },
        })
        
        if err := d.client.RegisterObject(sensorObj); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Object State Synchronization

Keep objects synchronized with device state:

```go
func (d *Driver) StartStateSynchronization() {
    // Poll device state every 60 seconds
    ticker := time.NewTicker(60 * time.Second)
    go func() {
        for range ticker.C {
            d.syncAllObjectStates()
        }
    }()
    
    // Also listen for device events (if supported)
    d.deviceClient.OnEvent(func(event DeviceEvent) {
        d.handleDeviceEvent(event)
    })
}

func (d *Driver) syncAllObjectStates() {
    deviceState, err := d.deviceClient.GetCurrentState()
    if err != nil {
        log.Printf("Failed to get device state: %v", err)
        return
    }
    
    // Update each object based on device state
    for objectID, object := range d.objects {
        if deviceData, exists := deviceState[objectID]; exists {
            object.SetState(deviceData.State)
            object.UpdateStateAttributes(deviceData.Attributes)
        }
    }
}
```

## Testing Objects

### Unit Testing Object Behavior

```go
func TestTemperatureSensor(t *testing.T) {
    // Create mock device
    mockDevice := &MockTemperatureDevice{
        temperature: 25.0,
    }
    
    // Create sensor object
    sensor := NewTemperatureSensor("test_device", "test_sensor", mockDevice)
    
    // Test initial state
    assert.Equal(t, "20.0", sensor.GetValue())
    
    // Test temperature update
    mockDevice.SetTemperature(30.0)
    sensor.UpdateFromDevice()
    
    assert.Equal(t, "30.0", sensor.GetValue())
}
```

### Integration Testing with Platform

```go
func TestObjectRegistration(t *testing.T) {
    // Create test client
    client := createTestClient(t)
    
    // Create and register object
    sensor := objects.NewSensorObject(objects.NewSensorObjectParams{
        Metadata: objects.ObjectMetadata{
            ObjectID: "test_sensor",
            Name:     "Test Temperature Sensor",
            Domain:   "temperature",
        },
    })
    
    err := client.RegisterObject(sensor)
    assert.NoError(t, err)
    
    // Verify object appears in platform
    objects, err := client.GetRegisteredObjects()
    assert.NoError(t, err)
    assert.Contains(t, objects, "test_sensor")
}
```

## Next Steps

Now that you understand objects thoroughly, explore these related topics:

- **[Configuration Handlers](handlers.md)** - Learn how to handle platform requests
- **[API Reference - Objects](api/objects/)** - Detailed reference for all object types
- **[Integration Guides](integrations/)** - See objects in action for specific device types
- **[Advanced State Management](advanced/state-management.md)** - Master complex state scenarios

## Quick Reference

### Common Object Creation Patterns

```go
// Sensor with setup function
sensor := objects.NewSensorObject(objects.NewSensorObjectParams{
    Metadata: metadata,
    SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
        sensor := obj.(objects.SensorObject)
        sensor.SetSensorType(objects.SensorObjectTypeNumber)
        sensor.SetUnitOfMeasurement("°C")
        return nil
    },
})

// Switch with action methods
sw := objects.NewSwitchObject(objects.NewSwitchObjectParams{
    Metadata: metadata,
    TurnOnMethod: turnOnHandler,
    TurnOffMethod: turnOffHandler,
})

// Camera with snapshot function
camera := objects.NewVideoChannelObject(objects.NewVideoChannelObjectProps{
    Metadata: metadata,
    StreamID: "rtsp://camera.local:554/stream1",
    SnapshotFn: captureSnapshot,
})
```

### State Update Patterns

```go
// Simple state change
object.SetState(objects.SENSOR_STATE_MEASUREMENT)

// Value update with attributes
sensor.SetValue("23.5")
sensor.UpdateStateAttributes(map[string]string{
    "unit": "°C",
    "battery": "85%",
})

// Batch update
attributes := map[string]string{
    "temperature": "23.5",
    "humidity": "65.0",
    "pressure": "1013.25",
}
object.UpdateStateAttributes(attributes)
```

Objects are the heart of your Netsocs driver. Master them, and you'll be able to create powerful, intuitive integrations for any IoT system.