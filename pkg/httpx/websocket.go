package httpx

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WebsocketDialer returns a websocket dialer that uses the same TLS
// configuration (and proxy settings) as the rest of the SDK, so wss:// endpoints
// behave exactly like https:// ones. Its TLS dial verifies the central once
// SecureCentral is set, as Transport does; DialWebsocket also refuses ws://
// to the central and adds the credentials.
func WebsocketDialer() *websocket.Dialer {
	d := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		TLSClientConfig:  TLSConfig(),
	}
	d.NetDialTLSContext = websocketDialTLS(d)
	return d
}

// websocketDialTLS makes the TLS handshake of d: the central's hosts get the
// configuration of SecureCentral, any other host d.TLSClientConfig (read at
// each dial). Through a proxy gorilla makes the handshake itself with
// d.TLSClientConfig.
func websocketDialTLS(d *websocket.Dialer) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr
		}
		cfg := tlsConfigForHost(host, d.TLSClientConfig)
		var nd net.Dialer
		raw, err := nd.DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		conn := tls.Client(raw, cfg)
		if err := conn.HandshakeContext(ctx); err != nil {
			_ = raw.Close()
			return nil, err
		}
		return conn, nil
	}
}
