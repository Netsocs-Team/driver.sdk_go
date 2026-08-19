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
package httpx

import (
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

// TLSConfig returns the TLS configuration used by the SDK. The returned value
// is a fresh copy, safe to mutate by the caller.
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
		client = &http.Client{Transport: transport}
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
	return t
}

// Transport returns the shared *http.Transport. Do not mutate it; use
// CloneTransport when a customized copy is needed.
func Transport() *http.Transport {
	initShared()
	return transport
}

// CloneTransport returns a private copy of the SDK transport, ready to be
// customized without affecting other callers.
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
	return &http.Client{Transport: Transport(), Timeout: timeout}
}

// Do performs the request with the shared client.
func Do(req *http.Request) (*http.Response, error) {
	return Client().Do(req)
}
