package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Netsocs-Team/driver.sdk_go/pkg/httpx"
)

func newTLSServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// The DriverHub is often served with a certificate signed by a private CA, so
// the shared client must not fail with "certificate signed by unknown
// authority".
func TestClientAcceptsPrivateCA(t *testing.T) {
	srv := newTLSServer(t)

	res, err := httpx.Client().Get(srv.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("got status %d, want 200", res.StatusCode)
	}
}

// A bare http.Client is the behaviour the SDK had before: it must fail, which is
// what this package exists to avoid.
func TestBareClientFailsOnPrivateCA(t *testing.T) {
	srv := newTLSServer(t)

	if _, err := (&http.Client{}).Get(srv.URL); err == nil {
		t.Fatal("expected the default client to reject the certificate")
	}
}

func TestRestyAcceptsPrivateCA(t *testing.T) {
	srv := newTLSServer(t)

	res, err := httpx.Resty().R().Get(srv.URL)
	if err != nil {
		t.Fatalf("resty request failed: %v", err)
	}
	if res.StatusCode() != http.StatusOK {
		t.Fatalf("got status %d, want 200", res.StatusCode())
	}
}

func TestWebsocketDialerUsesTheSDKTLSConfig(t *testing.T) {
	if !httpx.WebsocketDialer().TLSClientConfig.InsecureSkipVerify {
		t.Fatal("websocket dialer does not use the SDK TLS configuration")
	}
}
