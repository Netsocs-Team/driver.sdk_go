package httpx

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WebsocketDialer returns a websocket dialer that uses the same TLS
// configuration (and proxy settings) as the rest of the SDK, so wss:// endpoints
// behave exactly like https:// ones.
func WebsocketDialer() *websocket.Dialer {
	return &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		TLSClientConfig:  TLSConfig(),
	}
}
