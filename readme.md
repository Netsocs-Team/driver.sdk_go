# Netsocs Driver SDK for Go

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-latest-blue.svg)](./docs/)

The official Go SDK for developing Netsocs IoT device drivers. Build robust, production-ready drivers for cameras, access control systems, alarm panels, sensors, and custom IoT devices.

## 🚀 Quick Start

```bash
# Install the SDK
go get github.com/Netsocs-Team/driver.sdk_go

# Create a new driver from template
./scripts/new-driver.ps1 -Name my-driver -Module github.com/myorg/my-driver

# Build and run
cd my-driver
go mod tidy
go build -o my-driver
./my-driver
```

## 📋 Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Architecture Overview](#architecture-overview)
- [Quick Example](#quick-example)
- [Documentation](#documentation)
- [Driver Template](#driver-template)
- [Examples](#examples)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## ✨ Features

### Core SDK Capabilities
- **25+ Built-in Object Types**: Sensors, cameras, locks, alarms, GPS trackers, and more
- **70+ Configuration Handlers**: Pre-defined handlers for common device operations
- **Event System**: Dispatch events with images, videos, and custom properties
- **State Management**: Real-time state updates and attribute management
- **Connection Pooling**: Efficient device connection management
- **Type Safety**: Leverages Go's type system for robust development

### Supported Integrations
- **Video Surveillance**: IP cameras, NVRs, DVRs (Hikvision, Dahua, ONVIF)
- **Access Control**: Biometric readers, card readers, door controllers
- **Alarm Systems**: Security panels, zones, partitions
- **Environmental Monitoring**: Temperature, humidity, motion sensors
- **Cloud Services**: AWS SQS, webhooks, REST APIs

## 📚 Prerequisites

- **Go 1.21+**: [Download Go](https://golang.org/dl/)
- **Git**: For version control and dependency management
- **Netsocs Platform Access**: Driver credentials and platform endpoint
- **Device Documentation**: API specifications for target devices

## 🔧 Installation

### Method 1: Using Go Modules (Recommended)

```bash
# Initialize your driver project
mkdir my-netsocs-driver
cd my-netsocs-driver
go mod init github.com/myorg/my-netsocs-driver

# Install the SDK
go get github.com/Netsocs-Team/driver.sdk_go
```

### Method 2: Using the Template

```bash
# Clone the SDK repository
git clone https://github.com/Netsocs-Team/driver.sdk_go.git
cd driver.sdk_go

# Create a new driver from template
./scripts/new-driver.ps1 -Name my-driver -Module github.com/myorg/my-driver
```

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Netsocs Platform                         │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │   DriverHub     │    │   Web UI        │                │
│  │   (WebSocket)   │    │   (Actions)     │                │
│  └─────────┬───────┘    └─────────────────┘                │
└───────────┼─────────────────────────────────────────────────┘
            │ Configuration Requests
            │ State Updates & Events
            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Your Driver                              │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │ Config Handlers │    │    Objects      │                │
│  │ (Ping, Channels,│    │ (Sensors, Cams, │                │
│  │  Users, etc.)   │    │  Locks, etc.)   │                │
│  └─────────┬───────┘    └─────────┬───────┘                │
│            │                      │                        │
│            └──────────┬───────────┘                        │
│                       │                                    │
│  ┌─────────────────────▼─────────────────────┐              │
│  │           Device Manager                  │              │
│  │        (Connection Pooling)               │              │
│  └─────────────────────┬─────────────────────┘              │
└─────────────────────────┼─────────────────────────────────────┘
                          │ HTTP/TCP/WebSocket/SDK
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                Physical Devices                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Camera    │  │ Access Ctrl │  │ Alarm Panel │        │
│  │    NVR      │  │   Reader    │  │   Sensors   │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

### Key Components

1. **SDK Client**: Manages communication with the Netsocs platform
2. **Configuration Handlers**: Process platform requests for device operations
3. **Objects**: Represent devices and their capabilities (states, actions)
4. **Device Manager**: Handles connection pooling and device communication
5. **Events**: Notify the platform of significant occurrences

## 🎯 Quick Example

Here's a minimal temperature sensor driver:

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/Netsocs-Team/driver.sdk_go/pkg/client"
    "github.com/Netsocs-Team/driver.sdk_go/pkg/config"
    "github.com/Netsocs-Team/driver.sdk_go/pkg/objects"
)

func main() {
    // Initialize SDK client
    c, err := client.New()
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    // Register configuration handlers
    c.AddConfigHandler(config.ACTION_PING_DEVICE, handlePing)
    c.AddConfigHandler(config.REQUEST_CREATE_OBJECTS, handleCreateObjects(c))

    // Start listening for platform requests
    log.Println("Driver ready, listening for requests...")
    c.ListenConfig()
}

func handlePing(msg config.HandlerValue) (interface{}, error) {
    // Test device connectivity
    return map[string]interface{}{
        "status": true,
        "msg":    "Device is online",
    }, nil
}

func handleCreateObjects(c *client.NetsocsDriverClient) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        // Create temperature sensor
        sensor := objects.NewSensorObject(objects.NewSensorObjectParams{
            Metadata: objects.ObjectMetadata{
                ObjectID: "temp_sensor_01",
                Name:     "Living Room Temperature",
                Domain:   "temperature",
                DeviceID: msg.DeviceData.ID,
            },
            SetupFn: func(obj objects.RegistrableObject, oc objects.ObjectController) error {
                sensor := obj.(objects.SensorObject)
                sensor.SetSensorType(objects.SensorObjectTypeNumber)
                sensor.SetUnitOfMeasurement("°C")
                sensor.SetState(objects.SENSOR_STATE_MEASUREMENT)
                sensor.SetValue("20.0")
                
                // Start temperature updates
                go updateTemperature(sensor)
                return nil
            },
        })

        // Register with platform
        return nil, c.RegisterObject(sensor)
    }
}

func updateTemperature(sensor objects.SensorObject) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    temperature := 20.0
    for range ticker.C {
        temperature += (float64(time.Now().Unix()%3) - 1) * 0.5
        sensor.SetValue(fmt.Sprintf("%.1f", temperature))
    }
}
```

## 📖 Documentation

### Getting Started
- [Installation Guide](docs/installation.md)
- [Your First Driver](docs/first-driver.md)
- [Understanding Objects](docs/objects.md)
- [Configuration Handlers](docs/handlers.md)

### API Reference
- [Client API](docs/api/client.md)
- [Object Types](docs/api/objects/)
  - [Sensor Objects](docs/api/objects/sensor.md)
  - [Switch Objects](docs/api/objects/switch.md)
  - [Video Channel Objects](docs/api/objects/video-channel.md)
  - [Lock Objects](docs/api/objects/lock.md)
  - [Alarm Panel Objects](docs/api/objects/alarm-panel.md)
- [Configuration System](docs/api/config.md)
- [Events](docs/api/events.md)

### Advanced Topics
- [Device Connection Management](docs/advanced/device-management.md)
- [Event System Deep Dive](docs/advanced/events.md)
- [State Management](docs/advanced/state-management.md)
- [Error Handling](docs/advanced/error-handling.md)
- [Performance Optimization](docs/advanced/performance.md)
- [Security Best Practices](docs/advanced/security.md)

### Integration Guides
- [IP Camera Integration](docs/integrations/cameras.md)
- [Access Control Systems](docs/integrations/access-control.md)
- [Alarm Systems](docs/integrations/alarms.md)
- [Cloud Services](docs/integrations/cloud.md)

## 🏗️ Driver Template

The SDK includes a production-ready template with:

```
template/
├── main.go                      # Entry point with client initialization
├── go.mod                       # Go module definition
├── driver.netsocs.json.example  # Configuration template
├── config/
│   └── handlers.go             # Configuration request handlers
├── devices/
│   └── device_manager.go       # Device connection pooling
└── objects/
    ├── sensor_example.go       # Example sensor implementation
    └── switch_example.go       # Example switch implementation
```

### Using the Template

```bash
# Generate a new driver
./scripts/new-driver.ps1 -Name my-camera-driver -Module github.com/myorg/my-camera-driver

# Customize for your integration
cd my-camera-driver
# Edit config/handlers.go - implement your device API calls
# Edit objects/ - create objects for your device types
# Edit driver.netsocs.json - add your credentials

# Build and test
go mod tidy
go test ./...
go build -o my-camera-driver
./my-camera-driver
```

## 🔍 Examples

### Video Surveillance Driver

```go
// Register camera objects
camera := objects.NewVideoChannelObject(objects.NewVideoChannelObjectProps{
    Metadata: objects.ObjectMetadata{
        ObjectID: "camera_ch1",
        Name:     "Front Entrance Camera",
        Domain:   "camera",
        DeviceID: "nvr_001",
    },
    StreamID:    "rtsp://192.168.1.10:554/stream1",
    VideoEngine: "video_engine_01",
    PTZ:         true,
    SnapshotFn: func(vc objects.VideoChannelObject, oc objects.ObjectController,
                     payload objects.SnapshotActionPayload) (string, error) {
        // Capture and upload snapshot
        imageURL, err := captureSnapshot(vc.GetMetadata().ObjectID)
        return imageURL, err
    },
})
```

### Access Control Driver

```go
// Register reader object
reader := objects.NewReaderObject(objects.NewReaderObjectParams{
    Metadata: objects.ObjectMetadata{
        ObjectID: "reader_main_entrance",
        Name:     "Main Entrance Reader",
        Domain:   "access_control",
        DeviceID: "ac_panel_001",
    },
    SupportedCredentialTypes: []string{"card", "face", "fingerprint"},
})

// Dispatch access events
eventData := objects.Event{
    ObjectIDs: []string{"reader_main_entrance"},
    Properties: map[string]string{
        "user_id":     "12345",
        "credential":  "card",
        "result":      "granted",
        "door_id":     "main_door",
    },
}
client.DispatchEvent("access_control", "access_granted", eventData)
```

### Alarm System Driver

```go
// Register alarm panel
panel := objects.NewAlarmPanelObject(objects.NewAlarmPanelObjectProps{
    Metadata: objects.ObjectMetadata{
        ObjectID: "alarm_panel_main",
        Name:     "Main Security Panel",
        Domain:   "alarm",
        DeviceID: "panel_001",
    },
})

// Register zone sensors
for _, zone := range zones {
    sensor := objects.NewSensorObject(objects.NewSensorObjectParams{
        Metadata: objects.ObjectMetadata{
            ObjectID: fmt.Sprintf("zone_%s", zone.ID),
            Name:     zone.Name,
            Domain:   "alarm",
            DeviceID: "panel_001",
            ParentID: "alarm_panel_main",
        },
    })
    client.RegisterObject(sensor)
}
```

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/objects
```

### Integration Tests

```bash
# Test with mock device
go test -tags=integration ./tests/

# Test configuration handlers
go test ./config -run TestHandlers
```

### Example Test

```go
func TestPingHandler(t *testing.T) {
    handler := handlePingDevice(mockDeviceManager)
    
    msg := config.HandlerValue{
        DeviceData: config.DeviceData{
            IP:   "192.168.1.100",
            Port: 80,
        },
    }
    
    response, err := handler(msg)
    assert.NoError(t, err)
    
    result := response.(map[string]interface{})
    assert.True(t, result["status"].(bool))
}
```

## 🚀 Production Deployment

### Building for Production

```bash
# Build with version information
go build -ldflags="-X main.Version=1.0.0" -o my-driver

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o my-driver-linux
GOOS=windows GOARCH=amd64 go build -o my-driver-windows.exe
```

### Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o driver .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/driver .
COPY driver.netsocs.json .
CMD ["./driver"]
```

### Systemd Service (Linux)

```ini
[Unit]
Description=My Netsocs Driver
After=network.target

[Service]
Type=simple
User=netsocs
WorkingDirectory=/opt/my-driver
ExecStart=/opt/my-driver/my-driver
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/Netsocs-Team/driver.sdk_go.git
cd driver.sdk_go

# Install dependencies
go mod download

# Run tests
go test ./...

# Run linting
golangci-lint run
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: [https://docs.netsocs.com](https://docs.netsocs.com)
- **GitHub Issues**: [Report bugs or request features](https://github.com/Netsocs-Team/driver.sdk_go/issues)
- **Community**: Join our developer community for discussions and support

## 🏷️ Version History

- **v0.7.70**: Latest stable release with enhanced object types and improved error handling
- **v0.7.65**: Added cloud service integration support
- **v0.7.60**: Performance improvements and bug fixes
- See [CHANGELOG.md](CHANGELOG.md) for complete version history

---

**Ready to build your first driver?** Start with our [Installation Guide](docs/installation.md) and follow the [Quick Start Tutorial](docs/first-driver.md).