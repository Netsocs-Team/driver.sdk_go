package httpx

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
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
// certificate chain and the host name (see VerifyPinned for a central's own
// Caddy CA), and refuses plain http:// and ws://
// to them. mode is CAModePinned (caPEM is then the only trusted root) or
// CAModePublic (system roots). An empty mode removes the policy: every host
// is back to TLSConfig.
func SecureCentral(mode string, caPEM []byte, hosts ...string) error {
	if mode == "" {
		centralPolicy.Store(nil)
		dropSharedConnections()
		return nil
	}
	cfg := &tls.Config{MinVersion: tls.VersionTLS12}
	switch mode {
	case CAModePinned:
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caPEM) {
			return errors.New("httpx: the pinned CA file has no PEM certificate")
		}
		if PrivateRoot(caPEM) {
			cfg.InsecureSkipVerify = true //nolint:gosec // replaced by VerifyConnection
			cfg.VerifyConnection = VerifyChain(pool)
		} else {
			cfg.RootCAs = pool
		}
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
	dropSharedConnections()
	return nil
}

// dropSharedConnections closes the idle connections of the shared transport,
// which may have been opened under the previous policy.
func dropSharedConnections() {
	initShared()
	transport.CloseIdleConnections()
}

// VerifyChain checks the peer's chain against the pinned roots, not its
// name. It is used only for a central's own Caddy CA (PrivateRoot): that CA
// is generated per install and its keys never leave Caddy, so a chain to it
// already proves the peer is that Caddy, while the name in the certificate
// is often wrong for local installs (reached by IP with no SNI, Caddy names
// the container's address, 172.x, not the server's). Any other pinned CA,
// such as an organization PKI that signs other servers too, gets the
// standard verification with the name.
func VerifyChain(pool *x509.CertPool) func(tls.ConnectionState) error {
	return func(cs tls.ConnectionState) error {
		if len(cs.PeerCertificates) == 0 {
			return errors.New("httpx: the central sent no certificate")
		}
		inter := x509.NewCertPool()
		for _, c := range cs.PeerCertificates[1:] {
			inter.AddCert(c)
		}
		_, err := cs.PeerCertificates[0].Verify(x509.VerifyOptions{Roots: pool, Intermediates: inter})
		return err
	}
}

// PrivateRoot reports whether every certificate of caPEM is a self-signed
// root of a Caddy local authority ("Caddy Local Authority - <year> ECC
// Root"), the per-install CA of a central with no certificate of its own.
func PrivateRoot(caPEM []byte) bool {
	found := false
	for rest := caPEM; ; {
		var b *pem.Block
		b, rest = pem.Decode(rest)
		if b == nil {
			return found
		}
		c, err := x509.ParseCertificate(b.Bytes)
		if err != nil || !c.IsCA || c.Subject.String() != c.Issuer.String() ||
			!strings.HasPrefix(c.Subject.CommonName, "Caddy Local Authority") {
			return false
		}
		found = true
	}
}

// policyFor is the central policy when u points to a host of the central.
func policyFor(u *url.URL) *central {
	if u == nil {
		return nil
	}
	return policyForHost(u.Hostname())
}

// policyForHost is the central policy when host (a name or an IP, no port)
// is one of the central's.
func policyForHost(host string) *central {
	c := centralPolicy.Load()
	if c == nil || !c.hosts[strings.ToLower(strings.Trim(host, "[]"))] {
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
