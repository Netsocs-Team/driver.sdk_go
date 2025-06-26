package tools

import (
	"fmt"
	"strings"
)

// ConvertToWebSocketURL converts an HTTP/HTTPS URL to a WebSocket URL
// It handles protocol conversion (http -> ws, https -> wss) and ensures proper URL formatting
func ConvertToWebSocketURL(baseURL string, path string) string {
	// Start with the original URL
	url := baseURL

	// Convert HTTP/HTTPS to WebSocket protocol
	url = strings.ReplaceAll(url, "https://", "wss://")
	url = strings.ReplaceAll(url, "http://", "ws://")

	// If the URL doesn't have a protocol, assume it's HTTP and convert to WS
	if !strings.HasPrefix(url, "ws://") && !strings.HasPrefix(url, "wss://") {
		url = fmt.Sprintf("ws://%s", url)
	}

	// Remove trailing slash from base URL to avoid double slashes
	url = strings.TrimSuffix(url, "/")

	// Add the path if provided
	if path != "" {
		// Remove leading slash from path to avoid double slashes
		path = strings.TrimPrefix(path, "/")
		url = fmt.Sprintf("%s/%s", url, path)
	}

	return url
}
