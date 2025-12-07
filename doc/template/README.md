# Netsocs Driver Template

This template provides a complete boilerplate for developing Netsocs drivers. Use it as a starting point for integrating devices and systems with the Netsocs platform.

## Quick Start

### 1. Clone or Copy This Template

```bash
# Copy the template to a new project
cp -r template/ my-netsocs-driver/
cd my-netsocs-driver/
```

### 2. Configure Your Module

Update `go.mod` with your module name:

```go
module github.com/yourusername/my-netsocs-driver

go 1.21

require (
    github.com/Netsocs-Team/driver.sdk_go v0.0.0-latest
)
```

Then update all imports in the code:

```bash
# Replace "your-module-name" with your actual module name
find . -type f -name "*.go" -exec sed -i 's/your-module-name/github.com\/yourusername\/my-netsocs-driver/g' {} \;
```

### 3. Install Dependencies

```bash
go mod download
```

### 4. Configure Driver

Copy the example configuration and fill in your credentials:

```bash
cp driver.netsocs.json.example driver.netsocs.json
```

Edit `driver.netsocs.json` and replace placeholders with your actual values:
- `driver_key` - Your driver authentication key
- `driver_hub_host` - Platform URL
- `token` - Site authentication token
- `driver_id` - Your driver ID
- `site_id` - Your site ID

**Important**: Never commit `driver.netsocs.json` with real credentials to version control!

### 5. Build and Run

```bash
# Build
go build -o my-driver

# Run
./my-driver
```

You should see:

```
===========================================
 Starting Netsocs Driver
===========================================
Initializing client...
✓ Client initialized successfully
...
===========================================
 Driver ready, listening for requests...
===========================================
```

## Project Structure

```
my-netsocs-driver/
├── main.go                      # Entry point - client initialization
├── go.mod                       # Go module definition
├── driver.netsocs.json         # Configuration (gitignored)
├── driver.netsocs.json.example # Template configuration
├── config/
│   └── handlers.go             # Configuration request handlers
├── devices/
│   └── device_manager.go       # Device connection pooling
└── objects/
    ├── sensor_example.go       # Example sensor object
    └── switch_example.go       # Example switch object
```

## Customization Guide

### Adding a New Object Type

1. Create a new file in `objects/` (e.g., `my_object.go`)
2. Use SDK constructors like `objects.NewSensorObject()`, `objects.NewSwitchObject()`, etc.
3. Register in `main.go` → `registerObjects()`:

```go
myObj := objectsImpl.NewMyObject("obj_id", "device_id", deviceMgr)
c.RegisterObject(myObj)
```

### Adding Configuration Handlers

1. Edit `config/handlers.go`
2. Add your handler function:

```go
func handleMyAction(deviceMgr *devices.DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        // Your logic here
        return response, nil
    }
}
```

3. Register it in `RegisterAll()`:

```go
handlers := map[config.NetsocsConfigKey]config.FuncConfigHandler{
    config.MY_CONFIG_KEY: handleMyAction(deviceMgr),
}
```

4. Update `driver.netsocs.json` → `settings_available` array

### Implementing Device Communication

The template includes a `DeviceManager` for connection pooling. Customize `devices/device_manager.go`:

1. **Add your device SDK/client**:

```go
type Device struct {
    IP       string
    Port     int
    Username string
    Password string

    // Add your SDK here
    client *YourDeviceSDK
}
```

2. **Initialize connections** in `GetOrConnect()`:

```go
device.client, err = YourSDK.Connect(ip, port, username, password)
```

3. **Add device methods**:

```go
func (d *Device) DoSomething() error {
    return d.client.PerformAction()
}
```

4. **Use in handlers and objects**:

```go
device, err := deviceMgr.GetOrConnect(ip, port, username, password)
result, err := device.DoSomething()
```

### Dispatching Events

Add event types in `main.go` → `registerEventTypes()`:

```go
{
    Domain:             "camera",
    EventType:          "motion_detected",
    DisplayName:        "Motion Detected",
    DisplayDescription: "Camera detected motion",
    EventLevel:         "warning",
    Color:              "#FFA500",
    ShowColor:          true,
}
```

Dispatch events from anywhere in your code:

```go
eventData := objects.Event{
    ObjectIDs: []string{"camera_01"},
    ImageURLs: []string{imageURL},
    Properties: map[string]string{
        "confidence": "0.95",
    },
}

eventID, err := client.DispatchEvent("camera", "motion_detected", eventData)
```

## Testing

### Unit Testing

Create tests for your handlers and objects:

```go
// config/handlers_test.go
func TestPingDeviceHandler(t *testing.T) {
    deviceMgr := devices.NewDeviceManager()
    handler := handlePingDevice(deviceMgr)

    msg := config.HandlerValue{
        DeviceData: config.DeviceData{
            IP:   "192.168.1.100",
            Port: 80,
        },
    }

    response, err := handler(msg)
    if err != nil {
        t.Fatalf("Handler failed: %v", err)
    }

    // Assert response
}
```

Run tests:

```bash
go test ./...
```

### Manual Testing

1. **Check logs** for errors and warnings
2. **Verify in platform** that objects appear correctly
3. **Test actions** from the UI
4. **Monitor state updates** in real-time

## Common Integration Patterns

### Video Surveillance (IP Cameras, NVRs)

```go
// Register camera objects
camera := objects.NewVideoChannelObject(objects.NewVideoChannelObjectProps{
    Metadata: objects.ObjectMetadata{
        ObjectID: "camera_ch1",
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
    },
})
```

### Access Control

```go
// Register reader object
reader := objects.NewReaderObject(objects.NewReaderObjectParams{
    Metadata: objects.ObjectMetadata{
        ObjectID: "reader_01",
        Name:     "Main Entrance Reader",
        Domain:   "access_control",
        DeviceID: "ac_panel_001",
    },
    SupportedCredentialTypes: []string{"card", "face", "fingerprint"},
})
```

### Alarm Systems

```go
// Register alarm panel
panel := objects.NewAlarmPanelObject(objects.NewAlarmPanelObjectProps{
    Metadata: objects.ObjectMetadata{
        ObjectID: "alarm_panel_01",
        Name:     "Security Panel",
        Domain:   "alarm",
        DeviceID: "panel_001",
    },
})
```

## Deployment

### Building for Production

```bash
# Build with version info
go build -ldflags="-X main.Version=1.0.0" -o my-driver

# Or use build tags
go build -tags prod -o my-driver
```

### Running as a Service

#### Systemd (Linux)

Create `/etc/systemd/system/my-driver.service`:

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

Enable and start:

```bash
sudo systemctl enable my-driver
sudo systemctl start my-driver
sudo systemctl status my-driver
```

#### Docker

Create `Dockerfile`:

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

Build and run:

```bash
docker build -t my-netsocs-driver .
docker run -d --name my-driver --restart unless-stopped my-netsocs-driver
```

## Troubleshooting

### "Failed to create client"
- Ensure `driver.netsocs.json` exists in the current directory
- Verify JSON syntax is valid
- Check all required fields are present

### "Failed to register object"
- Ensure `ObjectID` is unique within your driver
- Check network connectivity to platform
- Verify credentials are correct

### "Connection refused"
- Ensure `driver_hub_host` URL is correct
- Check firewall settings
- Verify platform is accessible

### Objects not appearing in platform
- Check logs for registration errors
- Verify driver is activated in platform
- Ensure `site_id` and `driver_id` are correct

## Documentation

- [Full SDK Documentation](../docs/README.md)
- [Quick Start Guide](../docs/quick-start/01-installation.md)
- [API Reference](../docs/api-reference/client.md)
- [Configuration Handlers Guide](../docs/quick-start/04-configuration-handlers.md)

## Support

- **GitHub Issues**: [Report bugs or request features](https://github.com/Netsocs-Team/driver.sdk_go/issues)
- **Documentation**: [https://docs.netsocs.com](https://docs.netsocs.com)

## License

MIT License - See LICENSE file for details
