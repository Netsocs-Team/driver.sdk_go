// Package httpx centralizes the outbound HTTP/TLS configuration used by the
// SDK when it talks to the DriverHub (and to any other HTTPS endpoint).
//
// Netsocs installations frequently expose the DriverHub over HTTPS with a
// certificate issued by a private/corporate CA (or a self-signed one). The Go
// standard library only trusts the OS root store, so those deployments used to
// fail with:
//
//	tls: failed to verify certificate: x509: certificate signed by unknown authority
//
// The websocket connections of the SDK already accepted those certificates;
// every REST call, upload and event dispatch now goes through the transport
// built here, so all of them behave the same way and work out of the box, with
// no configuration.
//
// When driver.netsocs.json says how to trust the central (ca_mode, contract
// C1), SecureCentral switches the connections to the central's hosts
// (DriverHub, Keycloak) to full certificate and host name verification; the
// other hosts keep this configuration.
package httpx

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	once      sync.Once
	transport *http.Transport
	client    *http.Client
)

// TLSConfig returns the legacy TLS configuration of the SDK, used for every
// host that SecureCentral does not cover (TLSConfigFor picks the right one).
// The returned value is a fresh copy, safe to mutate by the caller.
func TLSConfig() *tls.Config {
	return &tls.Config{
		// The DriverHub is commonly served with a self-signed certificate or
		// one issued by a private CA that the host does not trust.
		//nolint:gosec // deliberate: on-premise DriverHubs use private CAs
		InsecureSkipVerify: true,
	}
}

func initShared() {
	once.Do(func() {
		transport = newTransport()
		client = &http.Client{Transport: authRoundTripper{router{}}}
	})
}

func newTransport() *http.Transport {
	var t *http.Transport
	if def, ok := http.DefaultTransport.(*http.Transport); ok {
		t = def.Clone()
	} else {
		t = &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			ExpectContinueTimeout: time.Second,
		}
	}
	t.TLSClientConfig = TLSConfig()
	t.TLSHandshakeTimeout = 15 * time.Second
	t.DialTLSContext = dialTLS(t)
	return t
}

// tlsConfigForHost is the TLS configuration of a connection to host: the
// verified one when host is the central's (SecureCentral), else base (the
// caller's own configuration, the legacy one by default). A fresh copy.
func tlsConfigForHost(host string, base *tls.Config) *tls.Config {
	var cfg *tls.Config
	if c := policyForHost(host); c != nil {
		cfg = c.tls.Clone()
		if base != nil {
			cfg.NextProtos = append([]string(nil), base.NextProtos...)
		}
	} else if base != nil {
		cfg = base.Clone()
	} else {
		cfg = TLSConfig()
	}
	if cfg.ServerName == "" {
		cfg.ServerName = host
	}
	return cfg
}

// dialTLS is the TLS dial of every transport the SDK builds, so that even
// Transport and CloneTransport, used directly by a driver, verify the
// central once ca_mode is set: its hosts get the configuration of
// SecureCentral, every other host t.TLSClientConfig (read at each dial, so a
// driver's own changes to its copy keep applying). A request through an
// HTTP proxy is the exception: net/http then makes the TLS handshake itself
// with t.TLSClientConfig; Client, NewClient, Resty and RoundTripper verify
// the central in every case.
func dialTLS(t *http.Transport) func(ctx context.Context, network, addr string) (net.Conn, error) {
	plain := (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr
		}
		cfg := tlsConfigForHost(host, t.TLSClientConfig)
		dial := t.DialContext
		if dial == nil {
			dial = plain
		}
		raw, err := dial(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		hctx := ctx
		if t.TLSHandshakeTimeout > 0 {
			var cancel context.CancelFunc
			hctx, cancel = context.WithTimeout(ctx, t.TLSHandshakeTimeout)
			defer cancel()
		}
		conn := tls.Client(raw, cfg)
		if err := conn.HandshakeContext(hctx); err != nil {
			_ = raw.Close()
			return nil, err
		}
		return conn, nil
	}
}

// Transport returns the shared *http.Transport. Its TLS dial verifies the
// central once SecureCentral is set (see dialTLS); it does not refuse plain
// http:// to the central nor add credentials (Client, NewClient, Resty and
// RoundTripper do). Do not mutate it; use CloneTransport when a customized
// copy is needed.
func Transport() *http.Transport {
	initShared()
	return transport
}

// CloneTransport returns a private copy of the SDK transport (legacy TLS for
// every host but the central's, see Transport), ready to be customized
// without affecting other callers.
func CloneTransport() *http.Transport {
	return newTransport()
}

// Client returns the shared *http.Client. It has no global timeout, mirroring
// http.DefaultClient; use NewClient for a bounded one. Do not mutate it.
func Client() *http.Client {
	initShared()
	return client
}

// NewClient returns a new *http.Client sharing the SDK transport, with the
// given timeout (zero means no timeout).
func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{Transport: authRoundTripper{router{}}, Timeout: timeout}
}

// Do performs the request with the shared client.
func Do(req *http.Request) (*http.Response, error) {
	return Client().Do(req)
}
