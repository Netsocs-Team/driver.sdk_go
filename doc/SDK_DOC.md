# Netsocs Driver SDK Documentation

Welcome to the official documentation for the Netsocs Driver SDK for Go. This SDK provides a comprehensive framework for building IoT device drivers that integrate with the Netsocs platform.

## What is the Netsocs Driver SDK?

The Netsocs Driver SDK enables developers to create drivers that connect IoT devices and systems to the Netsocs platform. Whether you're integrating IP cameras (like Hikvision, Dahua), access control systems, alarm panels, or custom IoT devices, this SDK provides the tools and patterns you need.

## Key Features

- **25+ Built-in Object Types**: Sensors, switches, cameras, locks, alarms, GPS trackers, and more
- **Event System**: Dispatch events with images, videos, and custom properties
- **Configuration Handlers**: 70+ predefined handlers for device operations
- **WebSocket & HTTP Communication**: Real-time action handling and state updates
- **Connection Pooling**: Efficient device connection management
- **Type-Safe**: Leverages Go's type system for robust driver development

## Prerequisites

Before you begin, ensure you have:

- **Go 1.21 or higher** installed
- Basic understanding of **Go programming**
- Familiarity with **IoT devices and protocols** (HTTP, RTSP, etc.)
- Access to the **Netsocs platform** with driver credentials

## Quick Links

### Getting Started
Start here if you're new to the SDK:

- [Installation](quick-start/01-installation.md) - Set up the SDK and create your configuration
- [Your First Driver](quick-start/02-first-driver.md) - Build a working driver in minutes
- [Understanding Objects](quick-start/03-understanding-objects.md) - Learn the core concepts
- [Configuration Handlers](quick-start/04-configuration-handlers.md) - Handle platform requests

### API Reference
Detailed documentation of all SDK components:

- [Client API](api-reference/client.md) - Main driver client interface
- [Object Types](api-reference/objects/overview.md) - All available object types
  - [Sensor](api-reference/objects/sensor.md)
  - [Switch](api-reference/objects/switch.md)
  - [Video Channel](api-reference/objects/video-channel.md)
  - [Lock](api-reference/objects/lock.md)
  - [Alarm Panel](api-reference/objects/alarm-panel.md)
  - [Reader](api-reference/objects/reader.md)
- [Configuration System](api-reference/config/overview.md)
  - [Handlers Reference](api-reference/config/handlers-reference.md)
- [Events](api-reference/events.md)

### Guides
In-depth guides for specific topics:

- [Object Types Guide](guides/object-types-guide.md) - Choosing and using object types
- [Event System Guide](guides/event-system-guide.md) - Dispatching events with media
- [State Management Guide](guides/state-management-guide.md) - Managing object states
- [Best Practices](guides/best-practices.md) - Recommended patterns and pitfalls

### Driver Template
Get started quickly with our generic driver template:

- [Generic Driver Template](../template/README.md) - Ready-to-use boilerplate code

## Architecture Overview

The Netsocs Driver SDK follows a clean architecture:

```
┌─────────────────┐
│  Your Driver    │
│  (Your Code)    │
└────────┬────────┘
         │
         │ Uses SDK
         ▼
┌─────────────────┐      WebSocket        ┌──────────────┐
│   SDK Client    │◄────────────────────► │  DriverHub   │
│                 │      (Actions)         │  (Platform)  │
│                 │                        │              │
│                 │◄────────────────────► │              │
│  Object Runner  │   HTTP (State Updates) │              │
└─────────────────┘                        └──────────────┘
         │
         │ Manages
         ▼
┌─────────────────┐
│    Objects      │
│ (Sensors, Cams, │
│  Locks, etc.)   │
└─────────────────┘
```

**Key Components:**

1. **Client**: Manages communication with the Netsocs platform
2. **Objects**: Represent devices and their capabilities (states, actions)
3. **Configuration Handlers**: Process platform requests for device operations
4. **Events**: Notify the platform of significant occurrences

## Common Use Cases

### Video Surveillance Integration
Integrate IP cameras with streaming, snapshots, PTZ control, and motion detection events.

- Use `VideoChannelObject` for camera channels
- Implement `GET_CHANNELS` config handler
- Dispatch motion detection events with snapshots

### Access Control Systems
Connect biometric readers, card readers, and door controllers.

- Use `ReaderObject` for credential readers
- Use `LockObject` for door locks
- Use `PersonObject` for managing people

### Alarm Systems
Integrate security alarm panels with partitions and zones.

- Use `AlarmPanelObject` for the panel
- Use `SensorObject` for zone sensors
- Implement arm/disarm config handlers

### Environmental Monitoring
Collect sensor data from IoT devices.

- Use `SensorObject` for measurements
- Set appropriate sensor types (temperature, humidity, etc.)
- Update values periodically

## Learning Path

We recommend following this learning path:

1. **Read the Quick Start Guide** (30 minutes)
   - Set up your environment
   - Build your first working driver
   - Understand core concepts

2. **Explore the Template** (15 minutes)
   - Review the generic driver template
   - Understand the project structure
   - See best practices in action

3. **Dive into API Reference** (as needed)
   - Reference client methods
   - Learn about specific object types
   - Understand configuration handlers

4. **Read the Guides** (as needed)
   - Master event dispatching
   - Optimize state management
   - Learn best practices

## Getting Help

- **GitHub Issues**: [Report bugs or request features](https://github.com/Netsocs-Team/driver.sdk_go/issues)
- **API Reference**: Check the [API documentation](api-reference/client.md) for detailed method signatures
- **Examples**: See the [template](../template/) for working code examples

## Contributing

We welcome contributions to the SDK and documentation. Please see our contributing guidelines for more information.

## License

This SDK is licensed under the MIT License. See LICENSE file for details.

---

**Ready to get started?** Begin with the [Installation Guide](quick-start/01-installation.md).
