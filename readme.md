# Netsocs Driver SDK

This is the official SDK for Netsocs Drivers, providing a robust framework for creating and managing IoT device drivers.

## Features

- Easy-to-use client interface for driver development
- Support for various object types (Sensors, Switches, etc.)
- Event management system
- Configuration handling
- Built-in logging capabilities

## Installation

```bash
go get github.com/Netsocs-Team/driver.sdk_go
```

## Quick Start

Here's a basic example of how to use the SDK:

```go
package main

import (
    "github.com/Netsocs-Team/driver.sdk_go/pkg/client"
    "github.com/Netsocs-Team/driver.sdk_go/pkg/config"
    "github.com/Netsocs-Team/driver.sdk_go/pkg/objects"
)

func main() {
    // Initialize the client
    client, err := client.New()
    if err != nil {
        panic(err)
    }

    // Create a sensor object
    sensorParams := objects.NewSensorObjectParams{}
    sensorParams.Metadata.Name = "temperature_sensor"
    sensorParams.Metadata.ObjectID = "temp_1"
    sensorParams.Metadata.Domain = "sensor"
    sensorParams.Metadata.DeviceID = "device_1"
    
    sensor := objects.NewSensorObject(sensorParams)
    
    // Register the sensor
    err = client.RegisterObject(sensor)
    if err != nil {
        panic(err)
    }
}
```

## Core Components

### Client
The `client` package provides the main interface for interacting with the Netsocs platform. It handles:
- Object registration
- Event dispatching
- Configuration management
- Connection handling

### Objects
The `objects` package defines various object types:
- `SensorObject`: For sensor devices
- `SwitchObject`: For switchable devices
- Custom objects can be created by implementing the `RegistrableObject` interface

### Events
The `event` package manages event handling:
- Event type registration
- Event dispatching
- Event filtering and processing

### Configuration
The `config` package handles:
- Driver configuration
- Configuration validation
- Configuration updates

## Advanced Usage

### Handling Events
```go
err := client.AddEventTypes([]objects.EventType{
    {
        Domain:             "custom",
        DisplayName:        "Custom Event",
        DisplayDescription: "Custom event description",
        EventType:          "custom_event",
        EventLevel:         "info",
        Color:              "#000000",
        ShowColor:          false,
        IsHidden:           false,
        Origin:             "driver",
    },
})
```

### Configuration Handlers
```go
client.AddConfigHandler(config.REQUEST_CREATE_OBJECTS, func(valueMessage config.HandlerValue) (interface{}, error) {
    // Handle configuration
    return nil, nil
})
```

## Best Practices

1. Always handle errors returned by SDK methods
2. Use appropriate object types for your devices
3. Implement proper cleanup in your drivers
4. Use the built-in logging system for debugging
5. Follow the event naming conventions

