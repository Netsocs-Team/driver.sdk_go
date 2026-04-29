# Installation Guide

This guide walks you through setting up the Netsocs Driver SDK and creating your first driver project.

## Prerequisites

Before you begin, ensure you have:

- **Go 1.21 or higher**: [Download Go](https://golang.org/dl/)
- **Git**: For version control and dependency management
- **Text editor or IDE**: VS Code, GoLand, or your preferred editor
- **Netsocs platform access**: Driver credentials from your platform administrator

## Step 1: Verify Go Installation

Confirm Go is properly installed and configured:

```bash
go version
# Should output: go version go1.21.x ...

echo $GOPATH
# Should show your Go workspace path

go env GOMODCACHE
# Should show your module cache location
```

If Go is not installed or configured properly, follow the [official Go installation guide](https://golang.org/doc/install).

## Step 2: Install the SDK

### Method A: Using Go Modules (Recommended)

Create a new driver project and install the SDK:

```bash
# Create project directory
mkdir my-netsocs-driver
cd my-netsocs-driver

# Initialize Go module
go mod init github.com/myorg/my-netsocs-driver

# Install the SDK
go get github.com/Netsocs-Team/driver.sdk_go

# Verify installation
go list -m github.com/Netsocs-Team/driver.sdk_go
```

### Method B: Using the Driver Template

Use our pre-built template for faster setup:

```bash
# Clone the SDK repository
git clone https://github.com/Netsocs-Team/driver.sdk_go.git
cd driver.sdk_go

# Create a new driver from template (Windows)
.\scripts\new-driver.ps1 -Name my-driver -Module github.com/myorg/my-driver

# Create a new driver from template (Linux/Mac)
./scripts/new-driver.sh -n my-driver -m github.com/myorg/my-driver

# Navigate to your new driver
cd my-driver
```

## Step 3: Verify SDK Installation

Create a simple test file to verify the SDK is working:

```go
// test-sdk.go
package main

import (
    "fmt"
    "github.com/Netsocs-Team/driver.sdk_go/pkg/client"
)

func main() {
    fmt.Println("Netsocs Driver SDK installed successfully!")
    
    // This will fail without proper configuration, but verifies imports work
    _, err := client.New()
    if err != nil {
        fmt.Printf("Expected error (no config file): %v\n", err)
    }
}
```

Run the test:

```bash
go run test-sdk.go
# Should output: Netsocs Driver SDK installed successfully!
# Expected error (no config file): ...
```

## Step 4: Configure Driver Credentials

### Create Configuration File

The SDK requires a configuration file named `driver.netsocs.json` in your project root:

```bash
# Create configuration file
touch driver.netsocs.json
```

### Configuration Template

Add the following structure to `driver.netsocs.json`:

```json
{
  "driver_key": "YOUR_DRIVER_KEY_HERE",
  "driver_hub_host": "https://your-platform.netsocs.com/api/netsocs/dh",
  "token": "YOUR_AUTH_TOKEN",
  "driver_id": "YOUR_DRIVER_ID",
  "site_id": "YOUR_SITE_ID",
  "name": "My First Driver",
  "version": "1.0.0",
  "driver_binary_filename": "my-driver",
  "documentation_url": "https://github.com/myorg/my-driver",
  "settings_available": [
    "actionPingDevice",
    "requestCreateObjects"
  ],
  "log_level": "info",
  "device_models_supported_all": true,
  "device_firmwares_supported_all": true
}
```

### Configuration Fields Reference

| Field | Required | Description |
|-------|----------|-------------|
| `driver_key` | ✅ | Authentication key for your driver (from platform admin) |
| `driver_hub_host` | ✅ | URL of the DriverHub API endpoint |
| `token` | ✅ | Authentication token for the site |
| `driver_id` | ✅ | Unique identifier for your driver |
| `site_id` | ✅ | Site identifier where the driver operates |
| `name` | ✅ | Human-readable name for your driver |
| `version` | ✅ | Driver version (semantic versioning recommended) |
| `driver_binary_filename` | ❌ | Name of the compiled binary |
| `documentation_url` | ❌ | URL to your driver's documentation |
| `settings_available` | ❌ | Array of configuration handlers your driver implements |
| `log_level` | ❌ | Logging level: `debug`, `info`, `warn`, `error` (default: `info`) |
| `device_models_supported_all` | ❌ | Whether driver supports all device models (default: `false`) |
| `device_firmwares_supported_all` | ❌ | Whether driver supports all firmware versions (default: `false`) |

### Getting Your Credentials

To obtain your credentials:

1. **Log in to the Netsocs platform**
2. **Navigate to the Drivers section**
3. **Create a new driver or select an existing one**
4. **Copy the required values:**
   - Driver Key
   - Driver ID  
   - Site ID
   - Auth Token
   - DriverHub Host URL

### Security Best Practices

⚠️ **Important Security Notes:**

- **Never commit `driver.netsocs.json` with real credentials to version control**
- Add `driver.netsocs.json` to your `.gitignore` file
- Use environment variables or secret management for production deployments
- Keep a template file (e.g., `driver.netsocs.json.example`) with placeholder values

Example `.gitignore` entry:

```gitignore
# Netsocs driver configuration with credentials
driver.netsocs.json

# Other common entries
*.log
*.exe
/dist/
/build/
```

## Step 5: Project Structure Setup

### Recommended Directory Structure

Organize your driver project with this structure:

```
my-netsocs-driver/
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── main.go                      # Entry point
├── driver.netsocs.json          # Configuration (gitignored)
├── driver.netsocs.json.example  # Template with placeholders
├── .gitignore                   # Git ignore rules
├── README.md                    # Project documentation
├── config/
│   └── handlers.go             # Configuration handlers
├── devices/
│   └── manager.go              # Device connection management
├── objects/
│   ├── sensors.go              # Sensor objects
│   └── cameras.go              # Camera objects (if applicable)
└── tests/
    ├── handlers_test.go        # Handler tests
    └── integration_test.go     # Integration tests
```

### Initialize Project Files

Create the basic project structure:

```bash
# Create directories
mkdir -p config devices objects tests

# Create basic files
touch main.go
touch config/handlers.go
touch devices/manager.go
touch README.md

# Create gitignore
cat > .gitignore << EOF
# Netsocs driver configuration
driver.netsocs.json

# Go build artifacts
*.exe
*.exe~
*.dll
*.so
*.dylib
/dist/
/build/

# Test binary, built with \`go test -c\`
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# Go workspace file
go.work
EOF

# Create configuration template
cp driver.netsocs.json driver.netsocs.json.example
# Edit the example file to replace real credentials with placeholders
```

## Step 6: Verify Complete Setup

### Create a Minimal Driver

Add this minimal driver code to `main.go`:

```go
package main

import (
    "log"
    
    "github.com/Netsocs-Team/driver.sdk_go/pkg/client"
    "github.com/Netsocs-Team/driver.sdk_go/pkg/config"
)

func main() {
    log.Println("Starting Netsocs driver...")
    
    // Initialize SDK client
    c, err := client.New()
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    
    log.Println("Client initialized successfully")
    
    // Add a simple ping handler
    c.AddConfigHandler(config.ACTION_PING_DEVICE, func(msg config.HandlerValue) (interface{}, error) {
        return map[string]interface{}{
            "status": true,
            "msg":    "Driver is running",
        }, nil
    })
    
    // Start listening for platform requests
    log.Println("Driver ready, listening for requests...")
    if err := c.ListenConfig(); err != nil {
        log.Fatalf("ListenConfig error: %v", err)
    }
}
```

### Build and Test

```bash
# Download dependencies
go mod tidy

# Build the driver
go build -o my-driver

# Run the driver (should connect to platform)
./my-driver
```

Expected output:
```
2024/01/15 10:30:00 Starting Netsocs driver...
2024/01/15 10:30:00 Client initialized successfully
2024/01/15 10:30:00 Driver ready, listening for requests...
```

If you see this output, your installation is complete! 🎉

## Troubleshooting

### Common Installation Issues

#### SDK Download Fails

```bash
# Error: cannot find module
go clean -modcache
go get github.com/Netsocs-Team/driver.sdk_go@latest

# If behind corporate firewall
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org
go get github.com/Netsocs-Team/driver.sdk_go
```

#### Authentication Errors

```bash
# Error: failed to create client
# Check your driver.netsocs.json file:

# 1. Verify JSON syntax
cat driver.netsocs.json | python -m json.tool

# 2. Check all required fields are present
grep -E "(driver_key|driver_hub_host|token|driver_id|site_id)" driver.netsocs.json

# 3. Verify no extra spaces or line breaks in credential strings
```

#### Connection Errors

```bash
# Error: connection refused
# 1. Verify DriverHub host URL is correct
curl -I https://your-platform.netsocs.com/api/netsocs/dh

# 2. Check firewall/network connectivity
telnet your-platform.netsocs.com 443

# 3. Verify your driver is activated in the platform
```

#### Module Path Issues

```bash
# Error: package not found
# Ensure your go.mod has the correct module name
head -1 go.mod

# Update imports if you changed the module name
find . -name "*.go" -exec sed -i 's/old-module-name/new-module-name/g' {} \;
```

### Getting Help

If you encounter issues not covered here:

1. **Check the [Troubleshooting Guide](deployment/troubleshooting.md)**
2. **Search [GitHub Issues](https://github.com/Netsocs-Team/driver.sdk_go/issues)**
3. **Create a new issue** with:
   - Go version (`go version`)
   - Operating system
   - Error messages (sanitized, no credentials)
   - Steps to reproduce

## Next Steps

Now that you have the SDK installed and configured:

1. **[Build Your First Driver](first-driver.md)** - Create a working temperature sensor driver
2. **[Understand Objects](objects.md)** - Learn about the core object system
3. **[Explore the Template](../template/README.md)** - Use the production-ready template
4. **[Browse Integration Guides](integrations/)** - Find guides for your device type

## Environment-Specific Setup

### Development Environment

```bash
# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Setup pre-commit hooks (optional)
cat > .git/hooks/pre-commit << EOF
#!/bin/bash
goimports -w .
golangci-lint run
go test ./...
EOF
chmod +x .git/hooks/pre-commit
```

### Production Environment

```bash
# Build for production
go build -ldflags="-X main.Version=1.0.0 -s -w" -o my-driver

# Create systemd service (Linux)
sudo tee /etc/systemd/system/my-driver.service > /dev/null << EOF
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
EOF

# Enable and start service
sudo systemctl enable my-driver
sudo systemctl start my-driver
```

Congratulations! You're now ready to build Netsocs drivers. Continue with [Your First Driver](first-driver.md) to create a complete working driver.