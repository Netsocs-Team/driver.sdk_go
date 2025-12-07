package devices

import (
	"fmt"
	"log"
	"sync"
)

// Device represents a connected device with its connection details
//
// This struct holds the connection information and any device-specific
// client or SDK instance. Customize this based on your device SDK.
type Device struct {
	IP       string
	Port     int
	Username string
	Password string

	// TODO: Add your device-specific client/SDK here
	// Examples:
	// client       *YourDeviceSDK
	// httpClient   *http.Client
	// wsConnection *websocket.Conn
	// apiClient    *YourAPIClient
}

// DeviceManager manages device connections with connection pooling
//
// This pattern avoids creating new connections for every request,
// improving performance and reducing load on devices.
type DeviceManager struct {
	devices map[string]*Device
	mu      sync.RWMutex
}

// NewDeviceManager creates a new device manager instance
func NewDeviceManager() *DeviceManager {
	return &DeviceManager{
		devices: make(map[string]*Device),
	}
}

// GetOrConnect retrieves an existing device connection or creates a new one
//
// This method is thread-safe and ensures only one connection per device.
// It uses IP:Port as the connection key.
func (dm *DeviceManager) GetOrConnect(ip string, port int, username, password string) (*Device, error) {
	key := fmt.Sprintf("%s:%d", ip, port)

	// Check if device already connected (read lock)
	dm.mu.RLock()
	if device, exists := dm.devices[key]; exists {
		dm.mu.RUnlock()
		log.Printf("Reusing existing connection to %s", key)
		return device, nil
	}
	dm.mu.RUnlock()

	// Create new connection (write lock)
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have created it)
	if device, exists := dm.devices[key]; exists {
		log.Printf("Connection to %s was created by another goroutine", key)
		return device, nil
	}

	log.Printf("Creating new connection to %s", key)

	// Create device instance
	device := &Device{
		IP:       ip,
		Port:     port,
		Username: username,
		Password: password,
	}

	// TODO: Initialize actual device connection
	//
	// Example implementations:
	//
	// 1. HTTP-based device:
	// device.httpClient = &http.Client{
	//     Timeout: 30 * time.Second,
	// }
	// device.baseURL = fmt.Sprintf("http://%s:%d", ip, port)
	//
	// 2. Custom SDK:
	// client, err := yourSDK.Connect(ip, port, username, password)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to connect: %w", err)
	// }
	// device.client = client
	//
	// 3. RTSP/Video device:
	// device.client = rtsp.NewClient(fmt.Sprintf("rtsp://%s:%s@%s:%d", username, password, ip, port))
	//
	// 4. WebSocket-based device:
	// wsURL := fmt.Sprintf("ws://%s:%d/api/ws", ip, port)
	// conn, err := websocket.Dial(wsURL, "", fmt.Sprintf("http://%s", ip))
	// if err != nil {
	//     return nil, err
	// }
	// device.wsConnection = conn

	// Test connection
	// err := device.Ping()
	// if err != nil {
	//     return nil, fmt.Errorf("device connection test failed: %w", err)
	// }

	// Store in manager
	dm.devices[key] = device
	log.Printf("Successfully connected to %s", key)

	return device, nil
}

// Remove removes a device connection from the pool
//
// Use this when a device goes offline or credentials change.
func (dm *DeviceManager) Remove(ip string, port int) {
	key := fmt.Sprintf("%s:%d", ip, port)

	dm.mu.Lock()
	defer dm.mu.Unlock()

	if device, exists := dm.devices[key]; exists {
		// TODO: Close device connection if needed
		// Examples:
		// device.client.Close()
		// device.wsConnection.Close()
		// device.httpClient.CloseIdleConnections()

		_ = device // Avoid unused variable error in template

		delete(dm.devices, key)
		log.Printf("Removed device connection: %s", key)
	}
}

// Cleanup closes all device connections
//
// Call this when the driver is shutting down to clean up resources.
func (dm *DeviceManager) Cleanup() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	log.Printf("Cleaning up %d device connections...", len(dm.devices))

	for key, device := range dm.devices {
		// TODO: Close device connections properly
		// Examples:
		// if device.client != nil {
		//     device.client.Close()
		// }
		// if device.wsConnection != nil {
		//     device.wsConnection.Close()
		// }

		_ = device // Avoid unused variable error in template
		log.Printf("  - Closed connection to %s", key)
	}

	dm.devices = make(map[string]*Device)
	log.Println("Cleanup complete")
}

// GetActiveConnections returns the number of active device connections
func (dm *DeviceManager) GetActiveConnections() int {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return len(dm.devices)
}

// Device Methods
//
// Add methods to the Device struct for interacting with your devices.
// These are examples - customize based on your device's API.

// Ping tests device connectivity
//
// TODO: Implement actual device ping logic
func (d *Device) Ping() error {
	// Example implementations:
	//
	// 1. HTTP-based:
	// resp, err := d.httpClient.Get(d.baseURL + "/api/status")
	// if err != nil {
	//     return err
	// }
	// resp.Body.Close()
	// return nil
	//
	// 2. SDK-based:
	// return d.client.TestConnection()
	//
	// 3. Simple TCP:
	// conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", d.IP, d.Port), 5*time.Second)
	// if err != nil {
	//     return err
	// }
	// conn.Close()
	// return nil

	log.Printf("Ping device %s:%d (mock - implement actual ping)", d.IP, d.Port)
	return nil
}

// GetChannels retrieves video channels from the device
//
// TODO: Implement actual channel retrieval logic
func (d *Device) GetChannels() ([]Channel, error) {
	// Example implementation:
	//
	// resp, err := d.httpClient.Get(d.baseURL + "/api/channels")
	// if err != nil {
	//     return nil, err
	// }
	// defer resp.Body.Close()
	//
	// var channels []Channel
	// if err := json.NewDecoder(resp.Body).Decode(&channels); err != nil {
	//     return nil, err
	// }
	//
	// return channels, nil

	log.Printf("GetChannels from %s:%d (mock - implement actual API call)", d.IP, d.Port)

	// Mock data for template
	return []Channel{
		{
			ID:        "1",
			Name:      "Channel 1",
			StreamURL: fmt.Sprintf("rtsp://%s:554/stream1", d.IP),
			Enabled:   true,
		},
		{
			ID:        "2",
			Name:      "Channel 2",
			StreamURL: fmt.Sprintf("rtsp://%s:554/stream2", d.IP),
			Enabled:   true,
		},
	}, nil
}

// SetRelayState controls a relay/switch on the device
//
// TODO: Implement actual relay control logic
func (d *Device) SetRelayState(relayID string, state bool) error {
	// Example implementation:
	//
	// payload := map[string]interface{}{
	//     "relay_id": relayID,
	//     "state":    state,
	// }
	//
	// jsonData, _ := json.Marshal(payload)
	// resp, err := d.httpClient.Post(d.baseURL+"/api/relay", "application/json", bytes.NewBuffer(jsonData))
	// if err != nil {
	//     return err
	// }
	// defer resp.Body.Close()
	//
	// if resp.StatusCode != 200 {
	//     return fmt.Errorf("failed to set relay state: %s", resp.Status)
	// }
	//
	// return nil

	stateStr := "OFF"
	if state {
		stateStr = "ON"
	}
	log.Printf("SetRelayState %s:%d relay=%s state=%s (mock - implement actual API call)",
		d.IP, d.Port, relayID, stateStr)

	return nil
}

// Channel represents a video channel from a camera or NVR
type Channel struct {
	ID        string
	Name      string
	StreamURL string
	Enabled   bool
}

// TODO: Add more device interaction methods as needed for your integration
//
// Examples:
// - func (d *Device) GetAlarmPartitions() ([]Partition, error)
// - func (d *Device) ArmPartition(partitionID string) error
// - func (d *Device) DisarmPartition(partitionID string) error
// - func (d *Device) GetSensorValue(sensorID string) (float64, error)
// - func (d *Device) CaptureSnapshot(channelID string) ([]byte, error)
// - func (d *Device) ControlPTZ(channelID, command string, value int) error
