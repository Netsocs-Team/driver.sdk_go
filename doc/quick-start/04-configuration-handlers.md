# Configuration Handlers

Configuration handlers allow your driver to respond to requests from the Netsocs platform. This guide explains what they are, how to implement them, and common use cases.

## What are Configuration Handlers?

Configuration handlers process requests from the platform for device operations. Examples include:

- **Ping a device** to test connectivity
- **Get video channels** from a camera or NVR
- **Retrieve alarm partitions** from a security panel
- **Add a person** to an access control system
- **Change video resolution** on a camera

When a user performs these actions in the platform UI, the platform sends a request to your driver via a configuration handler.

## How Configuration Handlers Work

```
Platform UI              Your Driver              Device
    │                         │                      │
    │──── Get Channels ──────▶│                      │
    │                         │───── API Call ──────▶│
    │                         │◀──── Channel List ───│
    │◀─── Return Channels ────│                      │
    │                         │                      │
```

1. **User action**: User clicks "Get Channels" in the UI
2. **Platform request**: Platform sends WebSocket message to your driver
3. **Handler execution**: Your `GET_CHANNELS` handler is called
4. **Device communication**: Handler communicates with the physical device
5. **Response**: Handler returns data to platform
6. **UI update**: Platform displays results to user

## Adding a Configuration Handler

Use `client.AddConfigHandler()` to register a handler:

```go
err := client.AddConfigHandler(config.ACTION_PING_DEVICE,
    func(msg config.HandlerValue) (interface{}, error) {
        // Your handler logic here
        return responseData, nil
    })
```

### Handler Function Signature

```go
func(msg config.HandlerValue) (interface{}, error)
```

**Parameters**:
- `msg` - Contains device data and request payload

**Returns**:
- `interface{}` - Response data (will be JSON marshaled automatically)
- `error` - Error if operation failed

### HandlerValue Structure

The `msg` parameter provides:

```go
type HandlerValue struct {
    DeviceData DeviceData  // Device connection info
    Value      string      // Request payload (JSON string)
}

type DeviceData struct {
    IP       string  // Device IP address
    Port     int     // Device port
    Username string  // Device username
    Password string  // Device password
    // ... additional fields
}
```

## Example: Ping Device Handler

The simplest handler - test device connectivity:

```go
import (
    "fmt"
    "net"
    "time"
    "github.com/Netsocs-Team/driver.sdk_go/pkg/config"
)

err := client.AddConfigHandler(config.ACTION_PING_DEVICE,
    func(msg config.HandlerValue) (interface{}, error) {
        // Get device IP from request
        deviceIP := msg.DeviceData.IP
        port := msg.DeviceData.Port

        // Test connection with timeout
        address := fmt.Sprintf("%s:%d", deviceIP, port)
        conn, err := net.DialTimeout("tcp", address, 5*time.Second)

        if err != nil {
            return map[string]interface{}{
                "status": false,
                "error":  true,
                "msg":    fmt.Sprintf("Device unreachable: %v", err),
            }, nil
        }

        conn.Close()

        return map[string]interface{}{
            "status": true,
            "error":  false,
            "msg":    "Device is online",
        }, nil
    })
```

## Example: Get Channels Handler

Retrieve video channels from a camera or NVR:

```go
import (
    "encoding/json"
    "github.com/Netsocs-Team/driver.sdk_go/pkg/config"
)

err := client.AddConfigHandler(config.GET_CHANNELS,
    func(msg config.HandlerValue) (interface{}, error) {
        // Extract device credentials
        deviceIP := msg.DeviceData.IP
        username := msg.DeviceData.Username
        password := msg.DeviceData.Password

        // Connect to device (example using hypothetical SDK)
        device, err := connectToDevice(deviceIP, username, password)
        if err != nil {
            return nil, fmt.Errorf("connection failed: %w", err)
        }
        defer device.Close()

        // Get channels from device
        channels, err := device.GetChannelList()
        if err != nil {
            return nil, fmt.Errorf("failed to get channels: %w", err)
        }

        // Format response for platform
        response := make([]map[string]interface{}, len(channels))
        for i, ch := range channels {
            response[i] = map[string]interface{}{
                "name":          ch.Name,
                "channelNumber": ch.ID,
                "rtspSource":    ch.StreamURL,
                "enabled":       ch.Enabled,
            }
        }

        return response, nil
    })
```

## Example: Set Video Resolution

Change camera resolution (with request payload):

```go
err := client.AddConfigHandler(config.SET_VIDEO_RESOLUTION,
    func(msg config.HandlerValue) (interface{}, error) {
        // Parse request payload
        var request struct {
            ChannelID  string `json:"channelId"`
            Resolution string `json:"resolution"`  // e.g., "1920x1080"
        }

        err := json.Unmarshal([]byte(msg.Value), &request)
        if err != nil {
            return nil, fmt.Errorf("invalid request: %w", err)
        }

        // Connect to device
        device, err := connectToDevice(
            msg.DeviceData.IP,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        if err != nil {
            return nil, err
        }
        defer device.Close()

        // Apply resolution change
        err = device.SetChannelResolution(request.ChannelID, request.Resolution)
        if err != nil {
            return nil, fmt.Errorf("failed to set resolution: %w", err)
        }

        return map[string]interface{}{
            "success": true,
            "message": fmt.Sprintf("Resolution changed to %s", request.Resolution),
        }, nil
    })
```

## Example: Get Alarm Partitions

Retrieve alarm system partitions:

```go
err := client.AddConfigHandler(config.GET_ALARM_PARTITIONS,
    func(msg config.HandlerValue) (interface{}, error) {
        device, err := connectToAlarmPanel(
            msg.DeviceData.IP,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        if err != nil {
            return nil, err
        }
        defer device.Close()

        partitions, err := device.GetPartitions()
        if err != nil {
            return nil, err
        }

        // Format response
        response := make([]map[string]interface{}, len(partitions))
        for i, partition := range partitions {
            response[i] = map[string]interface{}{
                "partition_number": partition.Number,
                "partition_name":   partition.Name,
                "armed":            partition.IsArmed,
                "alarm_state":      partition.AlarmState,
                "zones":            len(partition.Zones),
            }
        }

        return response, nil
    })
```

## Common Configuration Keys

The SDK provides 70+ configuration keys. Here are the most commonly used:

### Device Operations
- `ACTION_PING_DEVICE` - Test device connectivity
- `ACTION_RESTART_DEVICE` - Reboot the device
- `GET_EXTRA_DEVICE_FIELDS` - Get custom device properties
- `GET_DISCOVERED_DEVICES` - Device discovery

### Video Operations
- `GET_CHANNELS` - List video channels
- `SET_VIDEO_RESOLUTION` - Change resolution
- `GET_AVAILABLE_VIDEO_RESOLUTIONS` - List supported resolutions
- `SET_FLIP_VIDEO` - Flip video vertically
- `SET_MIRROR_VIDEO` - Mirror video horizontally
- `GET_RECORDING_RANGES` - Get recorded video time ranges

### Alarm System
- `GET_ALARM_PARTITIONS` - List alarm partitions
- `GET_ALARM_ZONES` - List alarm zones
- `GET_ALARM_USERS` - List alarm users
- `ACTION_ALARM_ARM_PARTITION` - Arm a partition
- `ACTION_ALARM_DISARM_PARTITION` - Disarm a partition
- `SET_ALARM_PARTITION_ZONE_BYPASS` - Bypass a zone

### Access Control
- `GET_ALL_PEOPLE_FROM_AC` - List all people
- `SET_ADD_PERSON_TO_AC` - Add a person
- `SET_DEL_PERSON_TO_AC` - Delete a person
- `SET_CARD_TO_PERSON_AC` - Assign card credential
- `SET_FACE_TO_PERSON_AC` - Assign face credential
- `SET_QR_TO_PERSON_AC` - Assign QR credential
- `SET_BLOCK_PERSON_TO_AC` - Block/unblock a person

### System Configuration
- `GET_USERS` - List device users
- `SET_USERS` - Update device users
- `GET_STORAGES` - List storage devices
- `GET_FTP_INFO` - Get FTP settings
- `SET_FTP_INFO` - Update FTP settings

### Special Handlers
- `REQUEST_CREATE_OBJECTS` - Platform requests driver to create objects
- `GET_EVENTS_AVAILABLE` - Return available event types

See [Handlers Reference](../api-reference/config/handlers-reference.md) for the complete list.

## Declaring Supported Handlers

In `driver.netsocs.json`, declare which handlers your driver implements:

```json
{
  "settings_available": [
    "actionPingDevice",
    "getChannels",
    "setVideoResolution",
    "getAlarmPartitions",
    "requestCreateObjects"
  ]
}
```

This helps the platform show only applicable actions in the UI.

## Error Handling

### Return Error for Failures

```go
err := client.AddConfigHandler(config.GET_CHANNELS,
    func(msg config.HandlerValue) (interface{}, error) {
        device, err := connect(msg.DeviceData.IP)
        if err != nil {
            // Return error - platform will show to user
            return nil, fmt.Errorf("connection failed: %w", err)
        }

        channels, err := device.GetChannels()
        if err != nil {
            return nil, fmt.Errorf("failed to retrieve channels: %w", err)
        }

        return channels, nil
    })
```

### Or Return Error in Response

For more control over error messages:

```go
err := client.AddConfigHandler(config.ACTION_PING_DEVICE,
    func(msg config.HandlerValue) (interface{}, error) {
        // ... ping logic ...

        if pingFailed {
            return map[string]interface{}{
                "status": false,
                "error":  true,
                "msg":    "Device unreachable - check IP address and network",
            }, nil  // Note: nil error, error info in response
        }

        return map[string]interface{}{
            "status": true,
            "error":  false,
            "msg":    "Device online",
        }, nil
    })
```

### Best Practices

1. **Descriptive Error Messages**: Help users understand what went wrong
   ```go
   // Good
   return nil, fmt.Errorf("channel 3 not found on device, available channels: 1-16")

   // Bad
   return nil, fmt.Errorf("error")
   ```

2. **Log Errors**: Use logging for debugging
   ```go
   log.Printf("Failed to connect to %s: %v", deviceIP, err)
   return nil, err
   ```

3. **Validate Input**: Check request payloads before processing
   ```go
   if request.ChannelID == "" {
       return nil, fmt.Errorf("channelId is required")
   }
   ```

4. **Timeout Long Operations**: Don't let handlers hang indefinitely
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()

   result, err := device.GetChannelsWithContext(ctx)
   ```

## Device Connection Management

To avoid reconnecting for every request, use a device manager pattern:

```go
type DeviceManager struct {
    devices map[string]*DeviceClient
    mu      sync.RWMutex
}

func (dm *DeviceManager) GetOrConnect(ip, username, password string) (*DeviceClient, error) {
    key := ip

    // Check if already connected
    dm.mu.RLock()
    if client, exists := dm.devices[key]; exists {
        dm.mu.RUnlock()
        return client, nil
    }
    dm.mu.RUnlock()

    // Connect to device
    dm.mu.Lock()
    defer dm.mu.Unlock()

    client, err := ConnectToDevice(ip, username, password)
    if err != nil {
        return nil, err
    }

    dm.devices[key] = client
    return client, nil
}
```

Then use in handlers:

```go
deviceMgr := NewDeviceManager()

err := client.AddConfigHandler(config.GET_CHANNELS,
    func(msg config.HandlerValue) (interface{}, error) {
        device, err := deviceMgr.GetOrConnect(
            msg.DeviceData.IP,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        if err != nil {
            return nil, err
        }

        return device.GetChannels()
    })
```

## Registering Multiple Handlers

Organize handlers in a separate function:

```go
func registerConfigHandlers(client *client.NetsocsDriverClient, deviceMgr *DeviceManager) error {
    handlers := map[config.NetsocsConfigKey]config.FuncConfigHandler{
        config.ACTION_PING_DEVICE:  handlePingDevice(deviceMgr),
        config.GET_CHANNELS:         handleGetChannels(deviceMgr),
        config.SET_VIDEO_RESOLUTION: handleSetVideoResolution(deviceMgr),
        config.GET_ALARM_PARTITIONS: handleGetAlarmPartitions(deviceMgr),
    }

    for key, handler := range handlers {
        if err := client.AddConfigHandler(key, handler); err != nil {
            return fmt.Errorf("failed to register %s: %w", key, err)
        }
    }

    return nil
}

// Helper function that returns a handler
func handlePingDevice(deviceMgr *DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        device, err := deviceMgr.GetOrConnect(...)
        // ... implementation ...
    }
}
```

## Testing Configuration Handlers

Test handlers work correctly before deploying:

```go
func TestPingDeviceHandler(t *testing.T) {
    handler := handlePingDevice(mockDeviceManager)

    // Create test message
    msg := config.HandlerValue{
        DeviceData: config.DeviceData{
            IP:       "192.168.1.100",
            Port:     80,
            Username: "admin",
            Password: "password",
        },
    }

    // Call handler
    response, err := handler(msg)

    // Assert results
    if err != nil {
        t.Fatalf("Handler returned error: %v", err)
    }

    result := response.(map[string]interface{})
    if result["status"] != true {
        t.Errorf("Expected status true, got %v", result["status"])
    }
}
```

## Next Steps

You now understand the core concepts of the Netsocs Driver SDK! Here's what to explore next:

- **Use the Template**: Start with the [Generic Driver Template](../../template/README.md)
- **API Reference**: Deep dive into [Client API](../api-reference/client.md) and [Object Types](../api-reference/objects/overview.md)
- **Advanced Topics**: Learn about [Events](../api-reference/events.md) and [State Management](../guides/state-management-guide.md)

## Quick Reference

### Handler Registration Pattern

```go
client.AddConfigHandler(config.KEY_NAME, func(msg config.HandlerValue) (interface{}, error) {
    // 1. Extract device info
    ip := msg.DeviceData.IP
    username := msg.DeviceData.Username
    password := msg.DeviceData.Password

    // 2. Parse request payload (if needed)
    var request RequestStruct
    json.Unmarshal([]byte(msg.Value), &request)

    // 3. Connect to device
    device, err := connect(ip, username, password)
    if err != nil {
        return nil, err
    }

    // 4. Perform operation
    result, err := device.DoSomething(request.Param)
    if err != nil {
        return nil, err
    }

    // 5. Return response
    return result, nil
})
```

### Common Imports

```go
import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/Netsocs-Team/driver.sdk_go/pkg/client"
    "github.com/Netsocs-Team/driver.sdk_go/pkg/config"
)
```
