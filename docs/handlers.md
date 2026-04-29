# Configuration Handlers

Configuration handlers are the communication bridge between the Netsocs platform and your driver. They process requests from the platform for device operations, configuration changes, and data retrieval. This comprehensive guide explains how to implement robust, production-ready handlers.

## What are Configuration Handlers?

Configuration handlers are functions that respond to specific requests from the Netsocs platform. When users interact with your driver through the platform UI, or when automations trigger actions, the platform sends requests to your driver via WebSocket. Your handlers process these requests and return responses.

### Request Flow

```
Platform UI → DriverHub → WebSocket → Your Handler → Device API → Response
```

1. **User Action**: User clicks "Get Channels" in the platform UI
2. **Platform Request**: Platform sends `GET_CHANNELS` request via WebSocket
3. **Handler Execution**: Your `GET_CHANNELS` handler is called
4. **Device Communication**: Handler communicates with the physical device
5. **Response**: Handler returns channel data to platform
6. **UI Update**: Platform displays results to user

## Handler Function Signature

All configuration handlers follow the same signature:

```go
func(msg config.HandlerValue) (interface{}, error)
```

### Parameters

**`msg config.HandlerValue`** contains:

```go
type HandlerValue struct {
    DeviceData DeviceData  // Device connection information
    Value      string      // Request payload (JSON string)
}

type DeviceData struct {
    ID          int                    // Device ID in platform
    Name        string                 // Device name
    IP          string                 // Device IP address
    Port        int                    // Device port
    Username    string                 // Device username
    Password    string                 // Device password
    IsSSL       bool                   // Whether to use SSL/TLS
    Extrafields map[string]interface{} // Additional custom fields
}
```

### Return Values

- **`interface{}`**: Response data (automatically JSON marshaled)
- **`error`**: Error if operation failed (displayed to user)

## Registering Handlers

### Single Handler Registration

```go
err := client.AddConfigHandler(config.ACTION_PING_DEVICE, func(msg config.HandlerValue) (interface{}, error) {
    // Handler implementation
    return response, nil
})
```

### Multiple Handler Registration (Recommended)

```go
func RegisterHandlers(client *client.NetsocsDriverClient, deviceMgr *DeviceManager) error {
    handlers := map[config.NetsocsConfigKey]config.FuncConfigHandler{
        config.ACTION_PING_DEVICE:    handlePingDevice(deviceMgr),
        config.REQUEST_CREATE_OBJECTS: handleCreateObjects(client, deviceMgr),
        config.GET_CHANNELS:           handleGetChannels(deviceMgr),
        config.GET_USERS:              handleGetUsers(deviceMgr),
    }
    
    for key, handler := range handlers {
        if err := client.AddConfigHandler(key, handler); err != nil {
            return fmt.Errorf("failed to register handler %s: %w", key, err)
        }
    }
    
    return nil
}
```

## Essential Handlers

Every driver should implement these core handlers:

### ACTION_PING_DEVICE

Tests device connectivity and credentials.

```go
func handlePingDevice(deviceMgr *DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        log.Printf("Pinging device %s:%d", msg.DeviceData.IP, msg.DeviceData.Port)
        
        // Get or create device connection
        device, err := deviceMgr.GetOrConnect(
            msg.DeviceData.IP,
            msg.DeviceData.Port,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        if err != nil {
            return map[string]interface{}{
                "status": false,
                "error":  true,
                "msg":    fmt.Sprintf("Connection failed: %v", err),
            }, nil
        }
        
        // Test device API
        if err := device.Ping(); err != nil {
            return map[string]interface{}{
                "status": false,
                "error":  true,
                "msg":    fmt.Sprintf("Device API error: %v", err),
            }, nil
        }
        
        return map[string]interface{}{
            "status": true,
            "error":  false,
            "msg":    "Device is online and responding",
        }, nil
    }
}
```

**When it's called**: User clicks "Test Connection" in device configuration

**Purpose**: Validate device reachability, credentials, and API compatibility

### REQUEST_CREATE_OBJECTS

Creates and registers objects for a device.

```go
func handleCreateObjects(client *client.NetsocsDriverClient, deviceMgr *DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        log.Printf("Creating objects for device: %s", msg.DeviceData.Name)
        
        deviceID := strconv.Itoa(msg.DeviceData.ID)
        
        // Validate extra fields if required
        if err := validateExtraFields(msg.DeviceData.Extrafields); err != nil {
            client.SetDeviceState(msg.DeviceData.ID, client.DeviceStateConfigurationFailure)
            return nil, fmt.Errorf("configuration error: %w", err)
        }
        
        // Connect to device
        device, err := deviceMgr.GetOrConnect(
            msg.DeviceData.IP,
            msg.DeviceData.Port,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        if err != nil {
            client.SetDeviceState(msg.DeviceData.ID, client.DeviceStateAuthenticationFailure)
            return nil, fmt.Errorf("device connection failed: %w", err)
        }
        
        // Discover device capabilities
        capabilities, err := device.GetCapabilities()
        if err != nil {
            return nil, fmt.Errorf("failed to discover device capabilities: %w", err)
        }
        
        // Create objects based on device capabilities
        objectCount := 0
        
        // Create sensor objects
        for _, sensor := range capabilities.Sensors {
            sensorObj := createSensorObject(deviceID, sensor)
            if err := client.RegisterObject(sensorObj); err != nil {
                return nil, fmt.Errorf("failed to register sensor %s: %w", sensor.Name, err)
            }
            objectCount++
        }
        
        // Create camera objects
        for _, camera := range capabilities.Cameras {
            cameraObj := createCameraObject(deviceID, camera)
            if err := client.RegisterObject(cameraObj); err != nil {
                return nil, fmt.Errorf("failed to register camera %s: %w", camera.Name, err)
            }
            objectCount++
        }
        
        // Register event types
        if err := client.AddEventTypes(getEventTypes()); err != nil {
            log.Printf("Failed to register event types: %v", err)
        }
        
        // Start background tasks
        go startDeviceMonitoring(device, client, deviceID)
        
        // Set device state to online
        client.SetDeviceState(msg.DeviceData.ID, client.DeviceStateOnline)
        
        log.Printf("Successfully created %d objects for device %s", objectCount, msg.DeviceData.Name)
        
        return map[string]interface{}{
            "success":        true,
            "objects_created": objectCount,
            "message":        fmt.Sprintf("Created %d objects", objectCount),
        }, nil
    }
}
```

**When it's called**: User clicks "Create Objects" after adding a device

**Purpose**: Discover device capabilities and create corresponding objects

## Handler Categories

### Device Operations

#### GET_EXTRA_DEVICE_FIELDS

Defines additional fields required for device configuration.

```go
func handleGetExtraDeviceFields() config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        return config.GetDeviceExtraFieldsResponse{
            &config.ExtraDeviceFieldsDefinition{
                Name:        "API_KEY",
                Description: "Device API key for authentication",
                Type:        config.ExtraFieldTypeString,
                Required:    true,
            },
            &config.ExtraDeviceFieldsDefinition{
                Name:        "REGION",
                Description: "Device region (US, EU, ASIA)",
                Type:        config.ExtraFieldTypeString,
                Required:    false,
                DefaultValue: "US",
            },
            &config.ExtraDeviceFieldsDefinition{
                Name:        "POLLING_INTERVAL",
                Description: "Polling interval in seconds",
                Type:        config.ExtraFieldTypeNumber,
                Required:    false,
                DefaultValue: "30",
            },
        }, nil
    }
}
```

#### ACTION_RESTART_DEVICE

Restarts the physical device.

```go
func handleRestartDevice(deviceMgr *DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        device, err := deviceMgr.GetOrConnect(
            msg.DeviceData.IP,
            msg.DeviceData.Port,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        if err != nil {
            return nil, err
        }
        
        if err := device.Restart(); err != nil {
            return nil, fmt.Errorf("failed to restart device: %w", err)
        }
        
        return map[string]interface{}{
            "success": true,
            "message": "Device restart initiated",
        }, nil
    }
}
```

### Video Operations

#### GET_CHANNELS

Retrieves available video channels from cameras or NVRs.

```go
func handleGetChannels(deviceMgr *DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        device, err := deviceMgr.GetOrConnect(
            msg.DeviceData.IP,
            msg.DeviceData.Port,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        if err != nil {
            return nil, err
        }
        
        channels, err := device.GetChannels()
        if err != nil {
            return nil, fmt.Errorf("failed to get channels: %w", err)
        }
        
        // Format response for platform
        response := make([]map[string]interface{}, len(channels))
        for i, ch := range channels {
            response[i] = map[string]interface{}{
                "name":          ch.Name,
                "channelNumber": ch.Number,
                "rtspSource":    ch.StreamURL,
                "enabled":       ch.Enabled,
                "resolution":    ch.Resolution,
                "fps":           ch.FPS,
            }
        }
        
        return response, nil
    }
}
```

#### SET_VIDEO_RESOLUTION

Changes video resolution for a specific channel.

```go
func handleSetVideoResolution(deviceMgr *DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        // Parse request payload
        var request struct {
            ChannelID  string `json:"channelId"`
            Resolution string `json:"resolution"`
        }
        
        if err := json.Unmarshal([]byte(msg.Value), &request); err != nil {
            return nil, fmt.Errorf("invalid request payload: %w", err)
        }
        
        // Validate input
        if request.ChannelID == "" {
            return nil, fmt.Errorf("channelId is required")
        }
        if request.Resolution == "" {
            return nil, fmt.Errorf("resolution is required")
        }
        
        // Connect to device
        device, err := deviceMgr.GetOrConnect(
            msg.DeviceData.IP,
            msg.DeviceData.Port,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        if err != nil {
            return nil, err
        }
        
        // Apply resolution change
        if err := device.SetChannelResolution(request.ChannelID, request.Resolution); err != nil {
            return nil, fmt.Errorf("failed to set resolution: %w", err)
        }
        
        return map[string]interface{}{
            "success": true,
            "message": fmt.Sprintf("Resolution changed to %s for channel %s", 
                request.Resolution, request.ChannelID),
        }, nil
    }
}
```

#### GET_RECORDING_RANGES

Gets available recording time ranges for playback.

```go
func handleGetRecordingRanges(deviceMgr *DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        var request struct {
            ChannelID string `json:"channelId"`
            StartDate string `json:"startDate"` // YYYY-MM-DD format
            EndDate   string `json:"endDate"`   // YYYY-MM-DD format
        }
        
        if err := json.Unmarshal([]byte(msg.Value), &request); err != nil {
            return nil, fmt.Errorf("invalid request payload: %w", err)
        }
        
        device, err := deviceMgr.GetOrConnect(
            msg.DeviceData.IP,
            msg.DeviceData.Port,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        if err != nil {
            return nil, err
        }
        
        ranges, err := device.GetRecordingRanges(request.ChannelID, request.StartDate, request.EndDate)
        if err != nil {
            return nil, fmt.Errorf("failed to get recording ranges: %w", err)
        }
        
        return ranges, nil
    }
}
```

### Access Control Operations

#### GET_ALL_PEOPLE_FROM_AC

Retrieves all people from access control system.

```go
func handleGetAllPeople(deviceMgr *DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        device, err := deviceMgr.GetOrConnect(
            msg.DeviceData.IP,
            msg.DeviceData.Port,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        if err != nil {
            return nil, err
        }
        
        people, err := device.GetAllPeople()
        if err != nil {
            return nil, fmt.Errorf("failed to get people: %w", err)
        }
        
        // Format response
        response := make([]map[string]interface{}, len(people))
        for i, person := range people {
            response[i] = map[string]interface{}{
                "id":          person.ID,
                "name":        person.Name,
                "email":       person.Email,
                "department":  person.Department,
                "active":      person.Active,
                "credentials": person.Credentials,
            }
        }
        
        return response, nil
    }
}
```

#### SET_ADD_PERSON_TO_AC

Adds a new person to the access control system.

```go
func handleAddPerson(deviceMgr *DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        var request struct {
            Name       string `json:"name"`
            Email      string `json:"email"`
            Department string `json:"department"`
        }
        
        if err := json.Unmarshal([]byte(msg.Value), &request); err != nil {
            return nil, fmt.Errorf("invalid request payload: %w", err)
        }
        
        // Validate required fields
        if request.Name == "" {
            return nil, fmt.Errorf("name is required")
        }
        
        device, err := deviceMgr.GetOrConnect(
            msg.DeviceData.IP,
            msg.DeviceData.Port,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        if err != nil {
            return nil, err
        }
        
        personID, err := device.AddPerson(request.Name, request.Email, request.Department)
        if err != nil {
            return nil, fmt.Errorf("failed to add person: %w", err)
        }
        
        return map[string]interface{}{
            "success":   true,
            "person_id": personID,
            "message":   fmt.Sprintf("Person %s added successfully", request.Name),
        }, nil
    }
}
```

### Alarm System Operations

#### GET_ALARM_PARTITIONS

Retrieves alarm system partitions.

```go
func handleGetAlarmPartitions(deviceMgr *DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        device, err := deviceMgr.GetOrConnect(
            msg.DeviceData.IP,
            msg.DeviceData.Port,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        if err != nil {
            return nil, err
        }
        
        partitions, err := device.GetAlarmPartitions()
        if err != nil {
            return nil, fmt.Errorf("failed to get partitions: %w", err)
        }
        
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
    }
}
```

#### ACTION_ALARM_ARM_PARTITION

Arms an alarm partition.

```go
func handleArmPartition(deviceMgr *DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        var request struct {
            PartitionNumber int    `json:"partitionNumber"`
            ArmMode         string `json:"armMode"` // "away", "home", "night"
            UserCode        string `json:"userCode"`
        }
        
        if err := json.Unmarshal([]byte(msg.Value), &request); err != nil {
            return nil, fmt.Errorf("invalid request payload: %w", err)
        }
        
        device, err := deviceMgr.GetOrConnect(
            msg.DeviceData.IP,
            msg.DeviceData.Port,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        if err != nil {
            return nil, err
        }
        
        if err := device.ArmPartition(request.PartitionNumber, request.ArmMode, request.UserCode); err != nil {
            return nil, fmt.Errorf("failed to arm partition: %w", err)
        }
        
        return map[string]interface{}{
            "success": true,
            "message": fmt.Sprintf("Partition %d armed in %s mode", 
                request.PartitionNumber, request.ArmMode),
        }, nil
    }
}
```

## Handler Best Practices

### Input Validation

Always validate input before processing:

```go
func handleSetVideoResolution(deviceMgr *DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        // 1. Validate device data
        if msg.DeviceData.IP == "" {
            return nil, fmt.Errorf("device IP is required")
        }
        
        // 2. Parse and validate JSON payload
        var request struct {
            ChannelID  string `json:"channelId"`
            Resolution string `json:"resolution"`
        }
        
        if err := json.Unmarshal([]byte(msg.Value), &request); err != nil {
            return nil, fmt.Errorf("invalid JSON payload: %w", err)
        }
        
        // 3. Validate required fields
        if request.ChannelID == "" {
            return nil, fmt.Errorf("channelId is required")
        }
        
        // 4. Validate field values
        validResolutions := []string{"1920x1080", "1280x720", "640x480"}
        if !contains(validResolutions, request.Resolution) {
            return nil, fmt.Errorf("invalid resolution: %s. Valid options: %v", 
                request.Resolution, validResolutions)
        }
        
        // Continue with processing...
    }
}
```

### Error Handling

Provide clear, actionable error messages:

```go
func handlePingDevice(deviceMgr *DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        device, err := deviceMgr.GetOrConnect(
            msg.DeviceData.IP,
            msg.DeviceData.Port,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        
        if err != nil {
            // Categorize errors for better user experience
            switch {
            case strings.Contains(err.Error(), "connection refused"):
                return map[string]interface{}{
                    "status": false,
                    "error":  true,
                    "msg":    fmt.Sprintf("Device unreachable at %s:%d. Check IP address and network connectivity.", 
                        msg.DeviceData.IP, msg.DeviceData.Port),
                }, nil
                
            case strings.Contains(err.Error(), "authentication failed"):
                return map[string]interface{}{
                    "status": false,
                    "error":  true,
                    "msg":    "Authentication failed. Check username and password.",
                }, nil
                
            case strings.Contains(err.Error(), "timeout"):
                return map[string]interface{}{
                    "status": false,
                    "error":  true,
                    "msg":    "Connection timeout. Device may be slow to respond or unreachable.",
                }, nil
                
            default:
                return map[string]interface{}{
                    "status": false,
                    "error":  true,
                    "msg":    fmt.Sprintf("Connection failed: %v", err),
                }, nil
            }
        }
        
        // Test device API
        if err := device.Ping(); err != nil {
            return map[string]interface{}{
                "status": false,
                "error":  true,
                "msg":    fmt.Sprintf("Device API error: %v", err),
            }, nil
        }
        
        return map[string]interface{}{
            "status": true,
            "error":  false,
            "msg":    "Device is online and responding",
        }, nil
    }
}
```

### Timeout Management

Always use timeouts for device operations:

```go
func handleGetChannels(deviceMgr *DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        // Create context with timeout
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        device, err := deviceMgr.GetOrConnect(
            msg.DeviceData.IP,
            msg.DeviceData.Port,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        if err != nil {
            return nil, err
        }
        
        // Use context for device operations
        channels, err := device.GetChannelsWithContext(ctx)
        if err != nil {
            if ctx.Err() == context.DeadlineExceeded {
                return nil, fmt.Errorf("operation timed out after 30 seconds")
            }
            return nil, fmt.Errorf("failed to get channels: %w", err)
        }
        
        return channels, nil
    }
}
```

### Device State Management

Update device states appropriately:

```go
func handleCreateObjects(client *client.NetsocsDriverClient, deviceMgr *DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        deviceID := msg.DeviceData.ID
        
        // Validate extra fields
        if apiKey, ok := msg.DeviceData.Extrafields["API_KEY"].(string); !ok || apiKey == "" {
            client.SetDeviceState(deviceID, client.DeviceStateConfigurationFailure)
            client.WriteLog(deviceID, "API_KEY is required but not provided")
            return nil, fmt.Errorf("API_KEY is required")
        }
        
        // Test connection
        device, err := deviceMgr.GetOrConnect(
            msg.DeviceData.IP,
            msg.DeviceData.Port,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        if err != nil {
            client.SetDeviceState(deviceID, client.DeviceStateAuthenticationFailure)
            client.WriteLog(deviceID, fmt.Sprintf("Authentication failed: %v", err))
            return nil, err
        }
        
        // Create objects
        objectCount, err := createDeviceObjects(client, device, deviceID)
        if err != nil {
            client.SetDeviceState(deviceID, client.DeviceStateConfigurationFailure)
            client.WriteLog(deviceID, fmt.Sprintf("Failed to create objects: %v", err))
            return nil, err
        }
        
        // Success - set device online
        client.SetDeviceState(deviceID, client.DeviceStateOnline)
        client.WriteLog(deviceID, fmt.Sprintf("Successfully created %d objects", objectCount))
        
        return map[string]interface{}{
            "success":        true,
            "objects_created": objectCount,
        }, nil
    }
}
```

### Logging

Add comprehensive logging for debugging:

```go
func handleGetChannels(deviceMgr *DeviceManager) config.FuncConfigHandler {
    return func(msg config.HandlerValue) (interface{}, error) {
        log.Printf("GET_CHANNELS request for device %s (%s:%d)", 
            msg.DeviceData.Name, msg.DeviceData.IP, msg.DeviceData.Port)
        
        startTime := time.Now()
        
        device, err := deviceMgr.GetOrConnect(
            msg.DeviceData.IP,
            msg.DeviceData.Port,
            msg.DeviceData.Username,
            msg.DeviceData.Password,
        )
        if err != nil {
            log.Printf("GET_CHANNELS failed for %s: connection error: %v", 
                msg.DeviceData.Name, err)
            return nil, err
        }
        
        channels, err := device.GetChannels()
        if err != nil {
            log.Printf("GET_CHANNELS failed for %s: API error: %v", 
                msg.DeviceData.Name, err)
            return nil, fmt.Errorf("failed to get channels: %w", err)
        }
        
        duration := time.Since(startTime)
        log.Printf("GET_CHANNELS completed for %s: %d channels retrieved in %v", 
            msg.DeviceData.Name, len(channels), duration)
        
        return formatChannelsResponse(channels), nil
    }
}
```

## Testing Handlers

### Unit Testing

```go
func TestPingDeviceHandler(t *testing.T) {
    // Create mock device manager
    mockDeviceMgr := &MockDeviceManager{
        devices: make(map[string]*MockDevice),
    }
    
    // Create handler
    handler := handlePingDevice(mockDeviceMgr)
    
    // Test successful ping
    t.Run("successful ping", func(t *testing.T) {
        msg := config.HandlerValue{
            DeviceData: config.DeviceData{
                IP:       "192.168.1.100",
                Port:     80,
                Username: "admin",
                Password: "password",
            },
        }
        
        // Mock successful device
        mockDeviceMgr.SetDevice("192.168.1.100:80", &MockDevice{
            pingResponse: nil, // No error = success
        })
        
        response, err := handler(msg)
        
        assert.NoError(t, err)
        result := response.(map[string]interface{})
        assert.True(t, result["status"].(bool))
        assert.False(t, result["error"].(bool))
    })
    
    // Test connection failure
    t.Run("connection failure", func(t *testing.T) {
        msg := config.HandlerValue{
            DeviceData: config.DeviceData{
                IP:   "192.168.1.999", // Invalid IP
                Port: 80,
            },
        }
        
        response, err := handler(msg)
        
        assert.NoError(t, err) // Handler should not return error for connection failures
        result := response.(map[string]interface{})
        assert.False(t, result["status"].(bool))
        assert.True(t, result["error"].(bool))
        assert.Contains(t, result["msg"].(string), "Connection failed")
    })
}
```

### Integration Testing

```go
func TestHandlerIntegration(t *testing.T) {
    // Setup test environment
    client := createTestClient(t)
    deviceMgr := devices.NewDeviceManager()
    
    // Register handlers
    err := RegisterHandlers(client, deviceMgr)
    assert.NoError(t, err)
    
    // Test ping handler
    pingResponse, err := client.TestHandler(config.ACTION_PING_DEVICE, config.HandlerValue{
        DeviceData: config.DeviceData{
            IP:       testDeviceIP,
            Port:     testDevicePort,
            Username: testUsername,
            Password: testPassword,
        },
    })
    
    assert.NoError(t, err)
    assert.True(t, pingResponse["status"].(bool))
}
```

## Complete Handler Examples

### Camera Driver Handlers

```go
func RegisterCameraHandlers(client *client.NetsocsDriverClient, deviceMgr *DeviceManager) error {
    handlers := map[config.NetsocsConfigKey]config.FuncConfigHandler{
        // Core handlers
        config.ACTION_PING_DEVICE:    handlePingDevice(deviceMgr),
        config.REQUEST_CREATE_OBJECTS: handleCreateObjects(client, deviceMgr),
        
        // Video handlers
        config.GET_CHANNELS:                     handleGetChannels(deviceMgr),
        config.SET_VIDEO_RESOLUTION:             handleSetVideoResolution(deviceMgr),
        config.GET_AVAILABLE_VIDEO_RESOLUTIONS:  handleGetAvailableResolutions(deviceMgr),
        config.GET_RECORDING_RANGES:             handleGetRecordingRanges(deviceMgr),
        
        // PTZ handlers
        config.ACTION_PTZ_UP:    handlePTZControl(deviceMgr, "up"),
        config.ACTION_PTZ_DOWN:  handlePTZControl(deviceMgr, "down"),
        config.ACTION_PTZ_LEFT:  handlePTZControl(deviceMgr, "left"),
        config.ACTION_PTZ_RIGHT: handlePTZControl(deviceMgr, "right"),
        config.ACTION_PTZ_ZOOM_IN:  handlePTZControl(deviceMgr, "zoom_in"),
        config.ACTION_PTZ_ZOOM_OUT: handlePTZControl(deviceMgr, "zoom_out"),
    }
    
    for key, handler := range handlers {
        if err := client.AddConfigHandler(key, handler); err != nil {
            return fmt.Errorf("failed to register handler %s: %w", key, err)
        }
    }
    
    return nil
}
```

### Access Control Driver Handlers

```go
func RegisterAccessControlHandlers(client *client.NetsocsDriverClient, deviceMgr *DeviceManager) error {
    handlers := map[config.NetsocsConfigKey]config.FuncConfigHandler{
        // Core handlers
        config.ACTION_PING_DEVICE:    handlePingDevice(deviceMgr),
        config.REQUEST_CREATE_OBJECTS: handleCreateObjects(client, deviceMgr),
        config.GET_EXTRA_DEVICE_FIELDS: handleGetExtraDeviceFields(),
        
        // People management
        config.GET_ALL_PEOPLE_FROM_AC: handleGetAllPeople(deviceMgr),
        config.SET_ADD_PERSON_TO_AC:   handleAddPerson(deviceMgr),
        config.SET_DEL_PERSON_TO_AC:   handleDeletePerson(deviceMgr),
        
        // Credential management
        config.SET_CARD_TO_PERSON_AC: handleAssignCard(deviceMgr),
        config.SET_FACE_TO_PERSON_AC: handleAssignFace(deviceMgr),
        config.SET_QR_TO_PERSON_AC:   handleAssignQR(deviceMgr),
        
        // Access logs
        config.GET_ACCESS_LOGS: handleGetAccessLogs(deviceMgr),
    }
    
    for key, handler := range handlers {
        if err := client.AddConfigHandler(key, handler); err != nil {
            return fmt.Errorf("failed to register handler %s: %w", key, err)
        }
    }
    
    return nil
}
```

## Handler Configuration

### Declaring Supported Handlers

In your `driver.netsocs.json`, declare which handlers your driver supports:

```json
{
  "settings_available": [
    "actionPingDevice",
    "requestCreateObjects",
    "getChannels",
    "setVideoResolution",
    "getRecordingRanges",
    "getAllPeopleFromAc",
    "setAddPersonToAc"
  ]
}
```

This helps the platform show only applicable actions in the UI.

## Next Steps

Now that you understand configuration handlers:

- **[API Reference - Configuration](api/config.md)** - Complete handler reference
- **[Integration Guides](integrations/)** - See handlers in action for specific devices
- **[Advanced Error Handling](advanced/error-handling.md)** - Master error handling patterns
- **[Testing Strategies](advanced/testing.md)** - Test your handlers thoroughly

## Quick Reference

### Handler Registration Pattern

```go
func RegisterHandlers(client *client.NetsocsDriverClient, deviceMgr *DeviceManager) error {
    handlers := map[config.NetsocsConfigKey]config.FuncConfigHandler{
        config.ACTION_PING_DEVICE: func(msg config.HandlerValue) (interface{}, error) {
            // 1. Extract device info
            // 2. Connect to device
            // 3. Test connectivity
            // 4. Return response
        },
    }
    
    for key, handler := range handlers {
        if err := client.AddConfigHandler(key, handler); err != nil {
            return fmt.Errorf("failed to register %s: %w", key, err)
        }
    }
    
    return nil
}
```

### Response Patterns

```go
// Success response
return map[string]interface{}{
    "success": true,
    "message": "Operation completed successfully",
    "data":    responseData,
}, nil

// Error response (return error)
return nil, fmt.Errorf("operation failed: %w", err)

// Ping response (success)
return map[string]interface{}{
    "status": true,
    "error":  false,
    "msg":    "Device is online",
}, nil

// Ping response (failure)
return map[string]interface{}{
    "status": false,
    "error":  true,
    "msg":    "Connection failed: timeout",
}, nil
```

Configuration handlers are the backbone of your driver's interaction with the platform. Implement them robustly, and your driver will provide a smooth, reliable user experience.