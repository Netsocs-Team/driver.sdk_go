# Netsocs Driver Generator Skill

A Claude Code skill that helps developers generate production-ready Netsocs IoT drivers for various device types.

## What This Skill Does

This skill automatically generates complete driver projects for the Netsocs platform, including:

- **IP Cameras** (Hikvision, Dahua, ONVIF, etc.)
- **Access Control Systems** (Biometric readers, card readers, etc.)
- **Alarm Systems** (Security panels, zones, partitions)
- **Generic IoT Devices** (Sensors, switches, relays, GPS trackers)

## When Claude Will Use This Skill

Claude automatically activates this skill when you:

- Ask to integrate a device with the Netsocs platform
- Request to create a new Netsocs driver
- Mention specific device types like cameras, access control, or alarm systems
- Need help with Netsocs driver development

## What Gets Generated

When you use this skill, Claude will create a complete driver project with:

```
driver-{device-name}/
├── main.go                      # Entry point with object registration
├── go.mod                       # Module definition
├── driver.netsocs.json.example  # Configuration template
├── .gitignore                   # Git ignore file
├── README.md                    # Custom setup instructions
├── config/
│   └── handlers.go              # Config handlers for your device
├── devices/
│   └── {device}_client.go       # Device SDK/API integration
└── objects/
    └── {object}.go              # Object implementations
```

## Example Usage

Simply ask Claude:

```
"I need to integrate Hikvision cameras with 16 channels"
```

Claude will:
1. Ask you questions about your integration needs
2. Generate a complete driver project
3. Add specific TODOs for device-specific implementation
4. Provide setup instructions and next steps

## Skill Structure

- **SKILL.md**: Main skill instructions and workflow
- **SDK_DOC.md**: Complete Netsocs SDK documentation
- **quick-start/**: Step-by-step guides for SDK usage
  - 01-installation.md
  - 02-first-driver.md
  - 03-understanding-objects.md
  - 04-configuration-handlers.md
- **template/**: Complete driver template with examples
  - main.go
  - config/handlers.go
  - devices/device_manager.go
  - objects/sensor_example.go
  - objects/switch_example.go

## Supported Integration Types

### 1. IP Cameras / Video Surveillance
- VideoChannelObject with streaming, PTZ, snapshots
- Motion detection events with images
- Multi-channel support for NVRs/DVRs

### 2. Access Control Systems
- ReaderObject for credential readers
- PersonObject for user management
- Face, card, fingerprint, QR credential support

### 3. Alarm Systems
- AlarmPanelObject for main panel
- SensorObject for zones
- Arm/disarm partition control

### 4. Generic IoT Devices
- SensorObject for environmental monitoring
- SwitchObject for relays and actuators
- GPS tracking and geofencing

## Tool Permissions

This skill has access to:
- **Read**: Read SDK templates and documentation
- **Write**: Create driver files
- **Grep/Glob**: Search for patterns in templates
- **Bash**: Create directory structures

## Tips for Best Results

1. **Be specific** about your device brand and model
2. **Mention the protocol** you'll use (HTTP, MQTT, SDK, etc.)
3. **List the features** you need (snapshots, PTZ, events, etc.)
4. **Share API documentation** if available

## Contributing

To improve this skill:

1. Update SKILL.md for workflow changes
2. Add new examples to quick-start guides
3. Enhance templates with better patterns
4. Document new object types in SDK_DOC.md

## Version

This skill is designed for the Netsocs Driver SDK for Go 1.21+.

---

**Need help?** Just ask Claude about integrating your device with Netsocs!
