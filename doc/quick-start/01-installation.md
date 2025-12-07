# Installation

This guide walks you through setting up the Netsocs Driver SDK and creating your driver configuration.

## Prerequisites

- **Go 1.21 or higher**: [Download Go](https://golang.org/dl/)
- **Git**: For cloning repositories and version control
- **Text editor or IDE**: VS Code, GoLand, or your preferred editor
- **Netsocs platform credentials**: Contact your platform administrator

## Step 1: Install the SDK

Add the Netsocs Driver SDK to your Go project:

```bash
go get github.com/Netsocs-Team/driver.sdk_go
```

This will download the SDK and add it to your `go.mod` file.

## Step 2: Verify Installation

Create a simple test file to verify the SDK is installed correctly:

```go
package main

import (
    "fmt"
    "github.com/Netsocs-Team/driver.sdk_go/pkg/client"
)

func main() {
    fmt.Println("Netsocs Driver SDK installed successfully!")
}
```

Run it:

```bash
go run main.go
```

You should see: `Netsocs Driver SDK installed successfully!`

## Step 3: Create Configuration File

The SDK requires a configuration file named `driver.netsocs.json` in your project root. This file contains credentials and settings for connecting to the Netsocs platform.

Create `driver.netsocs.json` with the following structure:

```json
{
  "driver_key": "YOUR_DRIVER_KEY_HERE",
  "driver_hub_host": "https://platform.netsocs.com/api/netsocs/dh",
  "token": "YOUR_AUTH_TOKEN",
  "driver_id": "YOUR_DRIVER_ID",
  "site_id": "YOUR_SITE_ID",
  "name": "My First Driver",
  "version": "1.0.0",
  "driver_binary_filename": "my-driver",
  "documentation_url": "https://github.com/yourusername/my-driver",
  "settings_available": [
    "actionPingDevice",
    "requestCreateObjects"
  ],
  "log_level": "info",
  "device_models_supported_all": true,
  "device_firmwares_supported_all": true
}
```

### Configuration Fields Explained

| Field | Required | Description |
|-------|----------|-------------|
| `driver_key` | Yes | Authentication key for your driver (provided by platform admin) |
| `driver_hub_host` | Yes | URL of the DriverHub API endpoint |
| `token` | Yes | Authentication token for the site |
| `driver_id` | Yes | Unique identifier for your driver |
| `site_id` | Yes | Site identifier where the driver will operate |
| `name` | Yes | Human-readable name for your driver |
| `version` | Yes | Driver version (semantic versioning recommended) |
| `driver_binary_filename` | No | Name of the compiled binary |
| `documentation_url` | No | URL to your driver's documentation |
| `settings_available` | No | Array of configuration handlers your driver implements |
| `log_level` | No | Logging level: `debug`, `info`, `warn`, `error` (default: `info`) |
| `device_models_supported_all` | No | Whether driver supports all device models (default: `false`) |
| `device_firmwares_supported_all` | No | Whether driver supports all firmware versions (default: `false`) |

### Getting Your Credentials

To obtain your credentials:

1. Log in to the Netsocs platform
2. Navigate to **Drivers** section
3. Create a new driver or select an existing one
4. Copy the following values:
   - Driver Key
   - Driver ID
   - Site ID
   - Auth Token
   - DriverHub Host URL

### Important Security Notes

- **Never commit `driver.netsocs.json` to version control** if it contains real credentials
- Add `driver.netsocs.json` to your `.gitignore` file
- Use environment variables or secret management for production deployments
- Keep a template file (e.g., `driver.netsocs.json.example`) with placeholder values

Example `.gitignore` entry:

```gitignore
# Netsocs driver configuration with credentials
driver.netsocs.json
```

## Step 4: Initialize Your Go Module

If you haven't already initialized a Go module for your driver project:

```bash
# Create a new directory for your driver
mkdir my-netsocs-driver
cd my-netsocs-driver

# Initialize Go module
go mod init github.com/yourusername/my-netsocs-driver

# Install the SDK
go get github.com/Netsocs-Team/driver.sdk_go
```

## Step 5: Project Structure

We recommend the following project structure:

```
my-netsocs-driver/
├── go.mod
├── go.sum
├── driver.netsocs.json      # Your configuration (gitignored)
├── driver.netsocs.json.example  # Template with placeholders
├── main.go                  # Entry point
├── config/
│   └── handlers.go          # Configuration handlers
├── devices/
│   └── manager.go           # Device connection management
└── objects/
    ├── sensors.go           # Sensor objects
    └── cameras.go           # Camera objects (if applicable)
```

## Next Steps

Now that you have the SDK installed and configured, you're ready to build your first driver!

Continue to: [Your First Driver](02-first-driver.md)

## Troubleshooting

### SDK Download Fails

If `go get` fails with connection errors:

```bash
# Try with explicit version
go get github.com/Netsocs-Team/driver.sdk_go@latest

# Or update your Go proxy settings
export GOPROXY=https://proxy.golang.org,direct
go get github.com/Netsocs-Team/driver.sdk_go
```

### Invalid Credentials

If you see authentication errors when running your driver:

1. Verify all credential fields in `driver.netsocs.json`
2. Ensure there are no extra spaces or line breaks in credential strings
3. Confirm the `driver_hub_host` URL is correct
4. Check that your driver is activated in the platform

### Module Not Found

If Go can't find the SDK module:

```bash
# Clear the module cache and retry
go clean -modcache
go get github.com/Netsocs-Team/driver.sdk_go
```

## Additional Resources

- [Go Modules Documentation](https://go.dev/blog/using-go-modules)
- [Netsocs Platform Documentation](https://docs.netsocs.com)
- [Generic Driver Template](../../template/README.md)
