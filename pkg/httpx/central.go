package httpx

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
)

// CA modes of the "auth" block of driver.netsocs.json (contract C1).
const (
	CAModePinned = "pinned" // only the root CA anchored when the site was registered
	CAModePublic = "public" // the system roots (the central has a public certificate)
)

// central is the TLS policy for the hosts of the central (DriverHub and
// Keycloak). Every other host keeps the legacy configuration of TLSConfig,
// so the HTTPS calls a driver makes to its devices do not change.
type central struct {
	hosts     map[string]bool // lower-case host names, no port
	tls       *tls.Config
	transport *http.Transport
}

var centralPolicy atomic.Pointer[central]

// SecureCentral makes every SDK connection to the given hosts verify the
// certificate chain and the host name, and refuses plain http:// and ws://
// to them. mode is CAModePinned (caPEM is then the only trusted root) or
// CAModePublic (system roots). An empty mode removes the policy: every host
// is back to TLSConfig.
func SecureCentral(mode string, caPEM []byte, hosts ...string) error {
	if mode == "" {
		centralPolicy.Store(nil)
		return nil
	}
	cfg := &tls.Config{MinVersion: tls.VersionTLS12}
	switch mode {
	case CAModePinned:
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caPEM) {
			return errors.New("httpx: the pinned CA file has no PEM certificate")
		}
		cfg.RootCAs = pool
	case CAModePublic:
	default:
		return fmt.Errorf("httpx: unknown ca_mode %q (want %q or %q)", mode, CAModePinned, CAModePublic)
	}
	c := &central{hosts: map[string]bool{}, tls: cfg}
	for _, h := range hosts {
		if h != "" {
			c.hosts[strings.ToLower(h)] = true
		}
	}
	if len(c.hosts) == 0 {
		return errors.New("httpx: no central host to secure")
	}
	c.transport = newTransport()
	c.transport.TLSClientConfig = cfg.Clone()
	centralPolicy.Store(c)
	return nil
}

// policyFor is the central policy when u points to a host of the central.
func policyFor(u *url.URL) *central {
	c := centralPolicy.Load()
	if c == nil || u == nil || !c.hosts[strings.ToLower(u.Hostname())] {
		return nil
	}
	return c
}

// ErrInsecureScheme: a plain http:// or ws:// request to the central once its
// TLS is verified.
var ErrInsecureScheme = errors.New("httpx: plain http/ws to the central is refused (ca_mode is set)")

func checkScheme(u *url.URL) error {
	if s := strings.ToLower(u.Scheme); s != "https" && s != "wss" {
		return fmt.Errorf("%w: %s://%s", ErrInsecureScheme, u.Scheme, u.Host)
	}
	return nil
}

// TLSConfigFor returns the TLS configuration the SDK uses for rawURL: the
// verified one for the central, TLSConfig for any other host. The returned
// value is a fresh copy, safe to mutate by the caller.
func TLSConfigFor(rawURL string) *tls.Config {
	u, err := url.Parse(rawURL)
	if err != nil {
		return TLSConfig()
	}
	if c := policyFor(u); c != nil {
		return c.tls.Clone()
	}
	return TLSConfig()
}

// router sends the requests to the central through the verified transport
// and every other request through the legacy one.
type router struct{}

func (router) RoundTrip(req *http.Request) (*http.Response, error) {
	c := policyFor(req.URL)
	if c == nil {
		return Transport().RoundTrip(req)
	}
	if err := checkScheme(req.URL); err != nil {
		if req.Body != nil {
			_ = req.Body.Close()
		}
		return nil, err
	}
	return c.transport.RoundTrip(req)
}

// RoundTripper returns the SDK transport without credentials: the central's
// hosts get the verified TLS configuration (see SecureCentral), every other
// host the legacy one. Use it for requests that must not carry the driver
// token (the Keycloak token endpoint itself).
func RoundTripper() http.RoundTripper { return router{} }
