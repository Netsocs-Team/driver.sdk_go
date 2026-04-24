package objects

import (
	"net/http"
	"strings"

	"github.com/Netsocs-Team/driver.sdk_go/pkg/tools"
)

// buildAudioStreamURL constructs the WebSocket URL for a DriversHub audio stream session.
// It strips any path suffix from hubHost so the URL always points to the root of the server.
func buildAudioStreamURL(hubHost, sessionID string) string {
	base := hubHost
	if strings.HasPrefix(hubHost, "http://") || strings.HasPrefix(hubHost, "https://") {
		// Keep only scheme + host:port, discard any path prefix that may be in hubHost
		withoutScheme := strings.SplitN(hubHost, "://", 2)[1]
		hostPort := strings.SplitN(withoutScheme, "/", 2)[0]
		scheme := strings.SplitN(hubHost, "://", 2)[0]
		base = scheme + "://" + hostPort
	}
	return tools.ConvertToWebSocketURL(base, "audio/stream/"+sessionID)
}

func wsDriverAuthHeader(driverKey string) http.Header {
	return http.Header{
		"Authorization": []string{driverKey},
	}
}
