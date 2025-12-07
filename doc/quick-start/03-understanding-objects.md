# Understanding Objects

Objects are the core abstraction in the Netsocs Driver SDK. This guide explains what objects are, how they work, and how to use them effectively.

## What is an Object?

An **object** represents a controllable or monitorable entity in your integrated system. Examples include:

- A temperature sensor
- A security camera
- A door lock
- An alarm panel
- A light switch
- A GPS tracker

Each object has:
- **Metadata**: Identification and classification (ID, name, type, domain)
- **States**: Current condition (e.g., "on", "off", "locked", "streaming")
- **Actions**: Operations that can be performed (e.g., "turn_on", "unlock", "snapshot")
- **State Attributes**: Additional key-value properties (e.g., temperature: "23.5", unit: "°C")

## The RegistrableObject Interface

All objects implement the `RegistrableObject` interface:

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

### Core Methods

#### 1. Setup(ObjectController) error

Called automatically after object registration. Use it for initialization:

```go
SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
    // Initialize connections
    // Set default states
    // Start background tasks
    return nil
}
```

**When it's called**: Immediately after `client.RegisterObject(obj)`

**Use cases**:
- Establish device connections
- Set initial state and attributes
- Start polling loops
- Subscribe to device events

**Example**:

```go
SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
    sensor := obj.(objects.SensorObject)

    // Set initial configuration
    sensor.SetSensorType(objects.SensorObjectTypeNumber)
    sensor.SetUnitOfMeasurement("°C")

    // Set initial state
    sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)
    sensor.SetValue("0.0")

    // Start background updates
    go pollDeviceAndUpdate(sensor)

    return nil
}
```

#### 2. GetAvailableStates() []string

Returns all possible states for this object type.

```go
func (s *sensorObject) GetAvailableStates() []string {
    return []string{
        "sensor.state.measurement",
        "sensor.state.total",
        "sensor.state.total_increasing",
    }
}
```

**Purpose**: The platform uses this to validate state transitions and display state history.

#### 3. GetAvailableActions() []ObjectAction

Returns all actions this object can perform.

```go
func (s *switchObject) GetAvailableActions() []ObjectAction {
    return []ObjectAction{
        {Action: "switch.action.turn_on", Domain: "switch"},
        {Action: "switch.action.turn_off", Domain: "switch"},
        {Action: "switch.action.toggle", Domain: "switch"},
    }
}
```

**Purpose**: The platform uses this to show available actions in the UI.

#### 4. RunAction(id, action string, payload []byte) (map[string]string, error)

Executes an action when triggered by a user or automation.

```go
func (s *switchObject) RunAction(id, action string, payload []byte) (map[string]string, error) {
    switch action {
    case "switch.action.turn_on":
        err := s.turnOn()
        return map[string]string{"status": "on"}, err
    case "switch.action.turn_off":
        err := s.turnOff()
        return map[string]string{"status": "off"}, err
    default:
        return nil, fmt.Errorf("unknown action: %s", action)
    }
}
```

**When it's called**: When a user clicks an action button in the UI or an automation is triggered.

#### 5. GetMetadata() ObjectMetadata

Returns the object's identification and classification.

```go
type ObjectMetadata struct {
    ObjectID string            // Unique ID within driver
    Name     string            // Display name
    Type     string            // Object type (sensor, switch, etc.)
    Domain   string            // Logical grouping
    I18n     map[string]string // Translations
    DeviceID string            // Parent device ID
    Tags     []string          // Searchable tags
    ParentID string            // Parent object ID (for hierarchies)
}
```

#### 6. SetState(state string) error

Updates the object's primary state.

```go
sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)
switch.SetState(objects.SWITCH_STATE_ON)
lock.SetState(objects.LOCK_STATE_LOCKED)
```

**Important**: State changes are sent to the platform via HTTP and persisted in the state history.

#### 7. UpdateStateAttributes(attributes map[string]string) error

Updates additional state properties.

```go
sensor.UpdateStateAttributes(map[string]string{
    "value":              "23.5",
    "unit_of_measurement": "°C",
    "battery_level":      "85",
})
```

## Object Lifecycle

Understanding the object lifecycle helps you know when each method is called:

```
┌──────────────────┐
│  Create Object   │  objects.NewSensorObject(params)
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Register Object  │  client.RegisterObject(obj)
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│   Setup() called │  Initialize, set defaults, start tasks
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Object Registered│  Visible in platform, ready for actions
└────────┬─────────┘
         │
         ▼
┌──────────────────────────────────────┐
│  Operation Phase                     │
│  ─────────────────                   │
│  • RunAction() when user triggers    │
│  • SetState() when condition changes │
│  • UpdateStateAttributes() as needed │
└──────────────────────────────────────┘
```

### Phase 1: Creation

You create an object using constructor functions:

```go
sensor := objects.NewSensorObject(params)
switch := objects.NewSwitchObject(params)
camera := objects.NewVideoChannelObject(props)
```

**At this point**: Object exists in memory but platform doesn't know about it.

### Phase 2: Registration

You register the object with the platform:

```go
err := client.RegisterObject(sensor)
```

**What happens**:
1. SDK sends object metadata to platform
2. Platform creates database entry
3. SDK calls `Setup()` automatically
4. Object becomes visible in UI

### Phase 3: Operation

Object is now active and can:
- **Receive actions** via `RunAction()`
- **Update states** via `SetState()`
- **Update attributes** via `UpdateStateAttributes()`

## Common Object Types

The SDK provides 25+ built-in object types. Here are the most common:

### SensorObject

For measurements and readings.

```go
sensor := objects.NewSensorObject(objects.NewSensorObjectParams{
    Metadata: objects.ObjectMetadata{
        ObjectID: "temp_01",
        Name:     "Temperature Sensor",
        Domain:   "temperature",
        DeviceID: "device_123",
    },
})
```

**Use for**: Temperature, humidity, motion, smoke, water leak, light level, etc.

**Key methods**:
- `SetValue(value string)` - Update reading
- `SetSensorType(type)` - Number, Text, Binary, Battery
- `SetUnitOfMeasurement(unit)` - °C, %, lux, etc.
- `Increment()` / `Decrement()` - For counters

**States**:
- `SENSOR_STATE_MEASUREMENT` - Current reading
- `SENSOR_STATE_TOTAL` - Cumulative value
- `SENSOR_STATE_TOTAL_INCREASING` - Ever-increasing counter

### SwitchObject

For on/off controllable devices.

```go
sw := objects.NewSwitchObject(objects.NewSwitchObjectParams{
    Metadata: objects.ObjectMetadata{
        ObjectID: "relay_01",
        Name:     "Living Room Light",
        Domain:   "light",
        DeviceID: "device_123",
    },
    TurnOnMethod: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
        // Send command to device to turn on
        return sendDeviceCommand("ON")
    },
    TurnOffMethod: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
        // Send command to device to turn off
        return sendDeviceCommand("OFF")
    },
})
```

**Use for**: Lights, relays, sirens, garage doors, etc.

**Actions**:
- `SWITCH_ACTION_TURN_ON` - Turn device on
- `SWITCH_ACTION_TURN_OFF` - Turn device off
- `SWITCH_ACTION_TOGGLE` - Toggle current state

**States**:
- `SWITCH_STATE_ON` - Device is on
- `SWITCH_STATE_OFF` - Device is off
- `SWITCH_STATE_UNKNOWN` - State unclear

### LockObject

For access control locks.

```go
lock := objects.NewLockObject(objects.NewLockObjectParams{
    Metadata: objects.ObjectMetadata{
        ObjectID: "door_lock_01",
        Name:     "Front Door Lock",
        Domain:   "lock",
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

**Use for**: Electronic locks, magnetic locks, electric strikes

**Actions**:
- `LOCK_ACTION_LOCK` - Secure the lock
- `LOCK_ACTION_UNLOCK` - Release the lock

### VideoChannelObject

For camera streams and video operations.

```go
camera := objects.NewVideoChannelObject(objects.NewVideoChannelObjectProps{
    Metadata: objects.ObjectMetadata{
        ObjectID: "camera_01_ch1",
        Name:     "Front Camera",
        Domain:   "camera",
        DeviceID: "nvr_001",
    },
    StreamID:    "rtsp://192.168.1.10:554/stream1",
    VideoEngine: "video_engine_01",
    PTZ:         true,
    SnapshotFn: func(vc objects.VideoChannelObject, oc objects.ObjectController,
                     payload objects.SnapshotActionPayload) (string, error) {
        // Capture snapshot, upload, return URL
        imageURL, err := captureAndUploadSnapshot()
        return imageURL, err
    },
})
```

**Use for**: IP cameras, DVRs, NVRs

**Actions**:
- `VIDEO_CHANNEL_ACTION_SNAPSHOT` - Capture image
- `VIDEO_CHANNEL_ACTION_VIDEOCLIP` - Record clip
- `VIDEO_CHANNEL_ACTION_PTZ_CONTROL` - Move camera

## Metadata Best Practices

### ObjectID

- Must be **unique within your driver**
- Use descriptive, consistent naming: `camera_01_ch1`, `temp_sensor_living_room`
- Include device ID or location for clarity
- Avoid special characters except `-` and `_`

```go
// Good
ObjectID: "device123_camera_ch1"
ObjectID: "temp_sensor_01"
ObjectID: "front_door_lock"

// Avoid
ObjectID: "obj1"
ObjectID: "sensor"
ObjectID: "camera ch1" // No spaces
```

### Domain

Groups related objects. Common domains:

- `temperature`, `humidity`, `motion`
- `camera`, `nvr`
- `lock`, `door`
- `alarm`, `security`
- `light`, `switch`

**Tip**: Use consistent domains across your drivers for better organization.

### Tags

Help users find and filter objects:

```go
Tags: []string{"indoor", "temperature", "living_room"}
Tags: []string{"outdoor", "camera", "perimeter"}
Tags: []string{"access_control", "biometric", "entrance"}
```

### Parent-Child Relationships

Use `DeviceID` and `ParentID` to create hierarchies:

```go
// NVR device
nvr := objects.NewVideoEngineObject(...)
nvr.Metadata.ObjectID = "nvr_001"

// Camera channels belong to NVR
camera1 := objects.NewVideoChannelObject(...)
camera1.Metadata.DeviceID = "nvr_001"  // Links to parent device
camera1.Metadata.ParentID = "nvr_001"  // Explicit parent relationship
```

## States vs. State Attributes

### States

- **Primary condition** of the object
- **Enumerated values** from `GetAvailableStates()`
- **Tracked in history** by the platform
- Examples: "on", "off", "locked", "streaming"

```go
sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)
```

### State Attributes

- **Additional properties** as key-value pairs
- **Flexible schema** (any string keys and values)
- **Updated frequently** without changing primary state
- Examples: temperature value, battery level, signal strength

```go
sensor.UpdateStateAttributes(map[string]string{
    "value":         "23.5",
    "battery_level": "85",
    "signal":        "strong",
})
```

### When to Use Each

**Use State** for:
- Primary operational condition
- Conditions users care about in history
- Triggering automations

**Use State Attributes** for:
- Measurement values
- Metadata (battery, signal, firmware version)
- Frequently changing data
- Additional context

**Example - Temperature Sensor**:

```go
// State: Type of measurement
sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)

// Attributes: Actual data
sensor.UpdateStateAttributes(map[string]string{
    "value":              "23.5",
    "unit_of_measurement": "°C",
    "battery_level":      "85%",
    "last_updated":       time.Now().Format(time.RFC3339),
})
```

## Custom Object Implementation

While the SDK provides built-in object types, you can implement `RegistrableObject` directly for custom needs:

```go
type myCustomObject struct {
    metadata   objects.ObjectMetadata
    controller objects.ObjectController
}

func (m *myCustomObject) Setup(oc objects.ObjectController) error {
    m.controller = oc
    // Custom initialization
    return nil
}

func (m *myCustomObject) GetAvailableStates() []string {
    return []string{"active", "inactive", "error"}
}

func (m *myCustomObject) GetAvailableActions() []objects.ObjectAction {
    return []objects.ObjectAction{
        {Action: "custom.action.do_something", Domain: "custom"},
    }
}

func (m *myCustomObject) RunAction(id, action string, payload []byte) (map[string]string, error) {
    // Handle custom actions
    return nil, nil
}

func (m *myCustomObject) GetMetadata() objects.ObjectMetadata {
    return m.metadata
}

func (m *myCustomObject) SetState(state string) error {
    return m.controller.SetState(m.metadata.ObjectID, state)
}

func (m *myCustomObject) UpdateStateAttributes(attrs map[string]string) error {
    return m.controller.UpdateStateAttributes(m.metadata.ObjectID, attrs)
}
```

**Note**: Built-in types are recommended for most use cases.

## Next Steps

Now that you understand objects, learn how to handle configuration requests:

- [Configuration Handlers](04-configuration-handlers.md)

Or dive deeper into specific object types:

- [Sensor API Reference](../api-reference/objects/sensor.md)
- [Switch API Reference](../api-reference/objects/switch.md)
- [Video Channel API Reference](../api-reference/objects/video-channel.md)
