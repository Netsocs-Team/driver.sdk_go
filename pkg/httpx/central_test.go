package httpx_test

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/Netsocs-Team/driver.sdk_go/internal/testpki"
	"github.com/Netsocs-Team/driver.sdk_go/pkg/httpx"
)

// newCentral serves HTTP and a websocket on /ws over TLS with cfg.
func newCentral(t *testing.T, cfg *tls.Config) *httptest.Server {
	t.Helper()
	up := websocket.Upgrader{}
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" {
			if c, err := up.Upgrade(w, r, nil); err == nil {
				_ = c.Close()
			}
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	srv.TLS = cfg
	srv.StartTLS()
	t.Cleanup(srv.Close)
	return srv
}

func secure(t *testing.T, mode string, caPEM []byte, hosts ...string) {
	t.Helper()
	if err := httpx.SecureCentral(mode, caPEM, hosts...); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = httpx.SecureCentral("", nil) })
}

// get calls srv over HTTP (shared client, resty, token transport) and over
// the websocket; all must agree.
func get(t *testing.T, srv *httptest.Server) error {
	t.Helper()
	var errs []error
	if res, err := httpx.Client().Get(srv.URL); err != nil {
		errs = append(errs, err)
	} else {
		res.Body.Close()
	}
	if _, err := httpx.Resty().R().Get(srv.URL); err != nil {
		errs = append(errs, err)
	}
	if res, err := (&http.Client{Transport: httpx.RoundTripper()}).Get(srv.URL); err != nil {
		errs = append(errs, err)
	} else {
		res.Body.Close()
	}
	if c, _, err := httpx.DialWebsocket("wss"+strings.TrimPrefix(srv.URL, "https")+"/ws", nil); err != nil {
		errs = append(errs, err)
	} else {
		_ = c.Close()
	}
	if len(errs) != 0 && len(errs) != 4 {
		t.Fatalf("the SDK connections disagree: %v", errs)
	}
	for _, err := range errs {
		if !strings.Contains(err.Error(), "x509") {
			t.Fatalf("failed for another reason than the certificate: %v", err)
		}
	}
	return errors.Join(errs...)
}

func TestPinnedCAAccepted(t *testing.T) {
	ca := testpki.NewCA(t, "central root")
	srv := newCentral(t, ca.Server(t, "127.0.0.1"))
	secure(t, httpx.CAModePinned, ca.PEM, "127.0.0.1")
	if err := get(t, srv); err != nil {
		t.Fatal(err)
	}
}

func TestPinnedRejectsAnotherCA(t *testing.T) {
	srv := newCentral(t, testpki.NewCA(t, "attacker").Server(t, "127.0.0.1"))
	secure(t, httpx.CAModePinned, testpki.NewCA(t, "central root").PEM, "127.0.0.1")
	if err := get(t, srv); err == nil {
		t.Fatal("a certificate from another CA was accepted")
	}
}

func TestPinnedRejectsTheWrongHostName(t *testing.T) {
	ca := testpki.NewCA(t, "central root")
	srv := newCentral(t, ca.Server(t, "central.example", "10.9.9.9"))
	secure(t, httpx.CAModePinned, ca.PEM, "127.0.0.1")
	if err := get(t, srv); err == nil {
		t.Fatal("a certificate for another host was accepted")
	}
}

// A central's own Caddy CA: reached by IP, Caddy names the container's
// address, so the name is not checked; the chain still is.
func TestPinnedCaddyRootAcceptsAnotherName(t *testing.T) {
	ca := testpki.NewCA(t, "Caddy Local Authority - 2026 ECC Root")
	srv := newCentral(t, ca.Server(t, "172.18.0.6"))
	secure(t, httpx.CAModePinned, ca.PEM, "127.0.0.1")
	if err := get(t, srv); err != nil {
		t.Fatalf("the central's Caddy certificate for its container address was refused: %v", err)
	}
	other := newCentral(t, testpki.NewCA(t, "Caddy Local Authority - 2026 ECC Root").Server(t, "127.0.0.1"))
	if err := get(t, other); err == nil {
		t.Fatal("another Caddy's CA was accepted")
	}
}

func TestPrivateRoot(t *testing.T) {
	if !httpx.PrivateRoot(testpki.NewCA(t, "Caddy Local Authority - 2026 ECC Root").PEM) {
		t.Fatal("a Caddy local root is private")
	}
	if httpx.PrivateRoot(testpki.NewCA(t, "Acme Corp Root CA").PEM) {
		t.Fatal("an organization CA keeps the name check")
	}
	both := append(testpki.NewCA(t, "Caddy Local Authority - 2026 ECC Root").PEM, testpki.NewCA(t, "Acme Corp Root CA").PEM...)
	if httpx.PrivateRoot(both) {
		t.Fatal("a bundle with an organization CA keeps the name check")
	}
}

func TestPublicRejectsAnUntrustedCertificate(t *testing.T) {
	srv := newCentral(t, testpki.NewCA(t, "private root").Server(t, "127.0.0.1"))
	secure(t, httpx.CAModePublic, nil, "127.0.0.1")
	if err := get(t, srv); err == nil {
		t.Fatal("a certificate outside the system roots was accepted")
	}
}

// Without ca_mode, and for hosts that are not the central (devices), the
// SDK keeps accepting private CAs as before.
func TestLegacyWithoutPolicyAndForOtherHosts(t *testing.T) {
	srv := newCentral(t, testpki.NewCA(t, "private root").Server(t, "127.0.0.1"))
	if err := get(t, srv); err != nil {
		t.Fatalf("no policy: %v", err)
	}
	secure(t, httpx.CAModePinned, testpki.NewCA(t, "central root").PEM, "central.example")
	if err := get(t, srv); err != nil {
		t.Fatalf("a device host got the central's policy: %v", err)
	}
	if !httpx.TLSConfigFor(srv.URL).InsecureSkipVerify || httpx.TLSConfigFor("https://central.example").InsecureSkipVerify {
		t.Fatal("TLSConfigFor picks the wrong configuration")
	}
}

func TestPlainHTTPToTheCentralIsRefused(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("the central got a plain request")
	}))
	t.Cleanup(srv.Close)
	secure(t, httpx.CAModePublic, nil, "127.0.0.1")

	if _, err := httpx.Client().Get(srv.URL); !errors.Is(err, httpx.ErrInsecureScheme) {
		t.Fatalf("http:// got %v", err)
	}
	if _, _, err := httpx.DialWebsocket("ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil); !errors.Is(err, httpx.ErrInsecureScheme) {
		t.Fatalf("ws:// got %v", err)
	}
}

func TestSecureCentralRejectsBadSettings(t *testing.T) {
	t.Cleanup(func() { _ = httpx.SecureCentral("", nil) })
	for name, err := range map[string]error{
		"pinned without PEM": httpx.SecureCentral(httpx.CAModePinned, []byte("not a pem"), "h"),
		"unknown mode":       httpx.SecureCentral("trust-me", nil, "h"),
		"no host":            httpx.SecureCentral(httpx.CAModePublic, nil),
	} {
		if err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
