# Netsocs Driver SDK Documentation

Welcome to the comprehensive documentation for the Netsocs Driver SDK for Go. This documentation will guide you through building production-ready IoT device drivers for the Netsocs platform.

## 📚 Documentation Structure

### Getting Started
Perfect for developers new to the Netsocs ecosystem.

- **[Installation](installation.md)** - Set up the SDK and configure your environment
- **[Your First Driver](first-driver.md)** - Build a working driver in 15 minutes
- **[Understanding Objects](objects.md)** - Core concepts and object lifecycle
- **[Configuration Handlers](handlers.md)** - Handle platform requests and actions

### API Reference
Detailed documentation for all SDK components.

- **[Client API](api/client.md)** - Main driver client interface and methods
- **[Object Types](api/objects/)** - Complete reference for all 25+ object types
  - [Sensor Objects](api/objects/sensor.md) - Temperature, humidity, motion sensors
  - [Switch Objects](api/objects/switch.md) - Relays, lights, controllable devices
  - [Video Channel Objects](api/objects/video-channel.md) - Cameras, NVRs, streaming
  - [Lock Objects](api/objects/lock.md) - Electronic locks, access control
  - [Alarm Panel Objects](api/objects/alarm-panel.md) - Security systems
  - [Reader Objects](api/objects/reader.md) - Biometric and card readers
  - [GPS Tracker Objects](api/objects/gps-tracker.md) - Location tracking
  - [And 18+ more object types...](api/objects/)
- **[Configuration System](api/config.md)** - 70+ configuration handlers reference
- **[Events API](api/events.md)** - Event types, dispatching, and media handling
- **[State Management](api/state-management.md)** - States vs attributes, updates

### Advanced Topics
In-depth guides for complex scenarios.

- **[Device Connection Management](advanced/device-management.md)** - Connection pooling and lifecycle
- **[Event System Deep Dive](advanced/events.md)** - Custom events, media handling, filtering
- **[Performance Optimization](advanced/performance.md)** - Scaling, memory management, concurrency
- **[Error Handling Patterns](advanced/error-handling.md)** - Robust error handling strategies
- **[Security Best Practices](advanced/security.md)** - Credential management, secure communication
- **[Testing Strategies](advanced/testing.md)** - Unit tests, integration tests, mocking
- **[Monitoring and Observability](advanced/monitoring.md)** - Logging, metrics, health checks

### Integration Guides
Step-by-step guides for specific device categories.

- **[IP Camera Integration](integrations/cameras.md)** - Hikvision, Dahua, ONVIF cameras
- **[Access Control Systems](integrations/access-control.md)** - Biometric readers, door controllers
- **[Alarm Systems](integrations/alarms.md)** - Security panels, zones, partitions
- **[Environmental Sensors](integrations/sensors.md)** - Temperature, humidity, air quality
- **[Cloud Services](integrations/cloud.md)** - AWS SQS, webhooks, REST APIs
- **[Custom Protocols](integrations/custom-protocols.md)** - TCP, UDP, serial communication

### Deployment & Operations
Production deployment and maintenance.

- **[Production Deployment](deployment/production.md)** - Docker, systemd, scaling
- **[Configuration Management](deployment/configuration.md)** - Environment variables, secrets
- **[Monitoring in Production](deployment/monitoring.md)** - Health checks, alerting
- **[Troubleshooting Guide](deployment/troubleshooting.md)** - Common issues and solutions
- **[Update Strategies](deployment/updates.md)** - Rolling updates, rollback procedures

## 🎯 Quick Navigation

### I want to...

**Build my first driver**
→ Start with [Installation](installation.md) → [Your First Driver](first-driver.md)

**Integrate a specific device type**
→ Check [Integration Guides](integrations/) for your device category

**Understand a specific object type**
→ Browse [Object Types Reference](api/objects/)

**Handle platform requests**
→ Read [Configuration Handlers](handlers.md) → [Config API Reference](api/config.md)

**Optimize performance**
→ See [Performance Optimization](advanced/performance.md)

**Deploy to production**
→ Follow [Production Deployment](deployment/production.md)

**Troubleshoot issues**
→ Check [Troubleshooting Guide](deployment/troubleshooting.md)

## 🏗️ Architecture Overview

Understanding the SDK architecture helps you build better drivers:

```
┌─────────────────────────────────────────────────────────────┐
│                    Netsocs Platform                         │
│                                                             │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │   DriverHub     │    │     Web UI      │                │
│  │  (WebSocket)    │    │   (Actions)     │                │
│  └─────────┬───────┘    └─────────────────┘                │
└───────────┼─────────────────────────────────────────────────┘
            │ Config Requests, State Updates, Events
            ▼
┌─────────────────────────────────────────────────────────────┐
│                  Your Driver (SDK)                          │
│                                                             │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │ Config Handlers │  │    Objects      │  │   Events    │ │
│  │ • Ping Device   │  │ • Sensors       │  │ • Motion    │ │
│  │ • Get Channels  │  │ • Cameras       │  │ • Access    │ │
│  │ • Add Person    │  │ • Locks         │  │ • Alarms    │ │
│  │ • Arm/Disarm    │  │ • Alarm Panels  │  │ • Custom    │ │
│  └─────────┬───────┘  └─────────┬───────┘  └─────────────┘ │
│            │                    │                          │
│            └────────────────────┼──────────────────────────┘
│                                 │
│  ┌─────────────────────────────┼─────────────────────────┐  │
│  │              Device Manager │                         │  │
│  │           (Connection Pool) │                         │  │
│  └─────────────────────────────┼─────────────────────────┘  │
└─────────────────────────────────┼─────────────────────────────┘
                                  │ HTTP/TCP/WebSocket/SDK
                                  ▼
┌─────────────────────────────────────────────────────────────┐
│                    Physical Devices                         │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Cameras   │  │ Access Ctrl │  │ Alarm Panel │        │
│  │    NVRs     │  │   Readers   │  │   Sensors   │        │
│  │   DVRs      │  │   Doors     │  │   Zones     │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

### Key Components

1. **SDK Client** (`pkg/client`)
   - Manages WebSocket connection to DriverHub
   - Handles authentication and heartbeat
   - Provides object registration and event dispatching

2. **Configuration Handlers** (`pkg/config`)
   - Process requests from the platform
   - 70+ pre-defined handler types
   - Custom handler implementation support

3. **Objects** (`pkg/objects`)
   - 25+ built-in object types
   - State and attribute management
   - Action handling framework

4. **Events** (`pkg/event`)
   - Event type registration
   - Event dispatching with media support
   - Custom property handling

## 🚀 Learning Path

We recommend following this learning path:

### Beginner (2-3 hours)
1. **[Installation](installation.md)** (15 minutes)
2. **[Your First Driver](first-driver.md)** (30 minutes)
3. **[Understanding Objects](objects.md)** (45 minutes)
4. **[Configuration Handlers](handlers.md)** (60 minutes)

### Intermediate (4-6 hours)
1. **[Client API Reference](api/client.md)** (60 minutes)
2. **[Object Types Deep Dive](api/objects/)** (120 minutes)
3. **[Event System](api/events.md)** (90 minutes)
4. **[Device Management](advanced/device-management.md)** (60 minutes)

### Advanced (8-12 hours)
1. **[Performance Optimization](advanced/performance.md)** (120 minutes)
2. **[Security Best Practices](advanced/security.md)** (90 minutes)
3. **[Testing Strategies](advanced/testing.md)** (120 minutes)
4. **[Production Deployment](deployment/production.md)** (180 minutes)

### Specialization (varies by integration)
Choose based on your target devices:
- **[Camera Integration](integrations/cameras.md)** for video surveillance
- **[Access Control](integrations/access-control.md)** for biometric systems
- **[Alarm Systems](integrations/alarms.md)** for security panels
- **[Cloud Services](integrations/cloud.md)** for cloud-based devices

## 📖 Code Examples

Throughout the documentation, you'll find practical examples:

### Simple Sensor Driver
```go
sensor := objects.NewSensorObject(objects.NewSensorObjectParams{
    Metadata: objects.ObjectMetadata{
        ObjectID: "temp_01",
        Name:     "Temperature Sensor",
        Domain:   "temperature",
    },
})
```

### Camera with PTZ Control
```go
camera := objects.NewVideoChannelObject(objects.NewVideoChannelObjectProps{
    StreamID: "rtsp://camera.local:554/stream1",
    PTZ:      true,
    SnapshotFn: captureSnapshot,
})
```

### Access Control Reader
```go
reader := objects.NewReaderObject(objects.NewReaderObjectParams{
    SupportedCredentialTypes: []string{"card", "face", "fingerprint"},
})
```

## 🔄 Updates and Versioning

This documentation is maintained alongside the SDK. Version compatibility:

- **Documentation v1.x**: SDK v0.7.60+
- **Latest**: Always reflects the current SDK version
- **Legacy**: Archived versions available in Git history

## 🤝 Contributing to Documentation

Found an error or want to improve the docs?

1. **Quick Fixes**: Create an issue describing the problem
2. **Content Contributions**: Submit a pull request with your changes
3. **New Guides**: Propose new integration guides or advanced topics

## 📞 Getting Help

- **GitHub Issues**: [Technical questions and bug reports](https://github.com/Netsocs-Team/driver.sdk_go/issues)
- **Documentation Issues**: Report unclear or missing documentation
- **Community**: Join our developer community for discussions

---

**Ready to start building?** Begin with the [Installation Guide](installation.md) and build your first driver with our [Quick Start Tutorial](first-driver.md).