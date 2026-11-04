package httpx

import (
	"net/http"
	"net/url"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

// AuthHook rewrites the credentials of a request to the DriverHub. It is
// installed by pkg/auth only when the driver runs with a Keycloak identity
// (the "auth" block of driver.netsocs.json and a DriverHub that accepts
// JWTs). Without it, every request leaves exactly as before.
type AuthHook func(u *url.URL, h http.Header)

var authHook atomic.Pointer[AuthHook]

// SetAuthHook installs (or, with nil, removes) the hook.
func SetAuthHook(h AuthHook) {
	if h == nil {
		authHook.Store(nil)
		return
	}
	authHook.Store(&h)
}

// ApplyAuth runs the hook, if any, on a request about to be sent to u.
func ApplyAuth(u *url.URL, h http.Header) {
	if p := authHook.Load(); p != nil {
		(*p)(u, h)
	}
}

// authRoundTripper applies the hook to every request of the SDK clients.
type authRoundTripper struct{ next http.RoundTripper }

func (a authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if authHook.Load() == nil {
		return a.next.RoundTrip(req)
	}
	// RoundTrippers must not modify the caller's request.
	r := req.Clone(req.Context())
	ApplyAuth(r.URL, r.Header)
	return a.next.RoundTrip(r)
}

// DialWebsocket dials with the SDK dialer after applying the hook to the
// handshake headers. To the central (SecureCentral) it verifies the
// certificate and refuses ws://.
func DialWebsocket(rawURL string, h http.Header) (*websocket.Conn, *http.Response, error) {
	if h == nil {
		h = http.Header{}
	}
	d := WebsocketDialer()
	if u, err := url.Parse(rawURL); err == nil {
		if c := policyFor(u); c != nil {
			if err := checkScheme(u); err != nil {
				return nil, nil, err
			}
			d.TLSClientConfig = c.tls.Clone()
		}
		ApplyAuth(u, h)
	}
	return d.Dial(rawURL, h)
}
