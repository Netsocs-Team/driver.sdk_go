package tools

import (
	"testing"
)

func TestConvertToWebSocketURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		path     string
		expected string
	}{
		{
			name:     "HTTPS to WSS with path",
			baseURL:  "https://example.com",
			path:     "api/ws",
			expected: "wss://example.com/api/ws",
		},
		{
			name:     "HTTP to WS with path",
			baseURL:  "http://example.com",
			path:     "api/ws",
			expected: "ws://example.com/api/ws",
		},
		{
			name:     "HTTPS to WSS without path",
			baseURL:  "https://example.com",
			path:     "",
			expected: "wss://example.com",
		},
		{
			name:     "HTTP to WS without path",
			baseURL:  "http://example.com",
			path:     "",
			expected: "ws://example.com",
		},
		{
			name:     "No protocol to WS with path",
			baseURL:  "example.com",
			path:     "api/ws",
			expected: "ws://example.com/api/ws",
		},
		{
			name:     "No protocol to WS without path",
			baseURL:  "example.com",
			path:     "",
			expected: "ws://example.com",
		},
		{
			name:     "Already WSS with path",
			baseURL:  "wss://example.com",
			path:     "api/ws",
			expected: "wss://example.com/api/ws",
		},
		{
			name:     "Already WS with path",
			baseURL:  "ws://example.com",
			path:     "api/ws",
			expected: "ws://example.com/api/ws",
		},
		{
			name:     "HTTPS with trailing slash to WSS with path",
			baseURL:  "https://example.com/",
			path:     "api/ws",
			expected: "wss://example.com/api/ws",
		},
		{
			name:     "HTTP with trailing slash to WS with path",
			baseURL:  "http://example.com/",
			path:     "api/ws",
			expected: "ws://example.com/api/ws",
		},
		{
			name:     "HTTPS to WSS with leading slash in path",
			baseURL:  "https://example.com",
			path:     "/api/ws",
			expected: "wss://example.com/api/ws",
		},
		{
			name:     "HTTP to WS with leading slash in path",
			baseURL:  "http://example.com",
			path:     "/api/ws",
			expected: "ws://example.com/api/ws",
		},
		{
			name:     "HTTPS with port to WSS with path",
			baseURL:  "https://example.com:8080",
			path:     "api/ws",
			expected: "wss://example.com:8080/api/ws",
		},
		{
			name:     "HTTP with port to WS with path",
			baseURL:  "http://example.com:8080",
			path:     "api/ws",
			expected: "ws://example.com:8080/api/ws",
		},
		{
			name:     "HTTPS with subdomain to WSS with path",
			baseURL:  "https://api.example.com",
			path:     "ws/v1/config_communication",
			expected: "wss://api.example.com/ws/v1/config_communication",
		},
		{
			name:     "HTTP with subdomain to WS with path",
			baseURL:  "http://api.example.com",
			path:     "ws/v1/config_communication",
			expected: "ws://api.example.com/ws/v1/config_communication",
		},
		{
			name:     "HTTPS with query parameters to WSS with path",
			baseURL:  "https://example.com?param=value",
			path:     "api/ws",
			expected: "wss://example.com?param=value/api/ws",
		},
		{
			name:     "Complex path with multiple segments",
			baseURL:  "https://devlabs001.netsocs.com/api/netsocs/dh",
			path:     "objects/ws",
			expected: "wss://devlabs001.netsocs.com/api/netsocs/dh/objects/ws",
		},
		{
			name:     "Complex path with query parameters",
			baseURL:  "https://devlabs001.netsocs.com/api/netsocs/dh",
			path:     "ws/v1/config_communication?site_id=1234567890&driver_id=1234567890",
			expected: "wss://devlabs001.netsocs.com/api/netsocs/dh/ws/v1/config_communication?site_id=1234567890&driver_id=1234567890",
		},
		{
			name:     "Empty base URL",
			baseURL:  "",
			path:     "api/ws",
			expected: "ws://api/ws",
		},
		{
			name:     "Empty base URL and path",
			baseURL:  "",
			path:     "",
			expected: "ws:/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToWebSocketURL(tt.baseURL, tt.path)
			if result != tt.expected {
				t.Errorf("ConvertToWebSocketURL(%q, %q) = %q, want %q", tt.baseURL, tt.path, result, tt.expected)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkConvertToWebSocketURL(b *testing.B) {
	baseURL := "https://devlabs001.netsocs.com/api/netsocs/dh"
	path := "objects/ws"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConvertToWebSocketURL(baseURL, path)
	}
}

func BenchmarkConvertToWebSocketURLNoPath(b *testing.B) {
	baseURL := "https://devlabs001.netsocs.com/api/netsocs/dh"
	path := ""

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConvertToWebSocketURL(baseURL, path)
	}
}
