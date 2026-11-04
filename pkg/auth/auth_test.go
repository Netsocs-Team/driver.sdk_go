package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Netsocs-Team/driver.sdk_go/pkg/httpx"
)

const driverKey = "UWawc/oDHUu5gG+Ipxd3g3wsRDu1"

func writeKey(t *testing.T, dir string) *ecdsa.PrivateKey {
	t.Helper()
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, _ := x509.MarshalECPrivateKey(k)
	if err := os.WriteFile(filepath.Join(dir, "driver.key"), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
	return k
}

// central fakes the DriverHub (/api/netsocs/dh) and Keycloak (/auth) of one
// central behind the same host.
type central struct {
	srv       *httptest.Server
	key       *ecdsa.PublicKey
	caps      int32 // 0: JWT accepted, 1: 404 (old DriverHub), 2: 503
	tokenFail int32 // token requests to fail with 503 before succeeding
	rejected  atomic.Bool
	tokens    int32
	lastHub   atomic.Value // http.Header of the last DriverHub request
}

func newCentral(t *testing.T, key *ecdsa.PublicKey) *central {
	return newCentralTLS(t, key, nil)
}

// newCentralTLS serves the central over TLS with cfg (plain HTTP if nil).
func newCentralTLS(t *testing.T, key *ecdsa.PublicKey, cfg *tls.Config) *central {
	c := &central{key: key}
	c.srv = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/netsocs/dh/auth/capabilities":
			switch atomic.LoadInt32(&c.caps) {
			case 1:
				http.NotFound(w, r)
			case 2:
				w.WriteHeader(http.StatusServiceUnavailable)
			default:
				_ = json.NewEncoder(w).Encode(map[string]any{"driver_jwt": true, "require_jwt": false, "reauth": true})
			}
		case r.URL.Path == "/auth/realms/netsocs/protocol/openid-connect/token":
			if r.Header.Get("Authorization") != "" {
				t.Errorf("token request carries Authorization %q", r.Header.Get("Authorization"))
			}
			if atomic.AddInt32(&c.tokenFail, -1) >= 0 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			_ = r.ParseForm()
			if c.rejected.Load() || !c.validAssertion(t, r.Form.Get("client_assertion"), r.Form.Get("client_id")) {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"invalid_client"}`))
				return
			}
			n := atomic.AddInt32(&c.tokens, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "jwt-" + string(rune('0'+n)), "expires_in": 300})
		case strings.HasPrefix(r.URL.Path, "/api/netsocs/dh/"):
			c.lastHub.Store(r.Header.Clone())
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	if cfg != nil {
		c.srv.TLS = cfg
		c.srv.StartTLS()
	} else {
		c.srv.Start()
	}
	t.Cleanup(c.srv.Close)
	return c
}

func (c *central) validAssertion(t *testing.T, a, clientID string) bool {
	parts := strings.Split(a, ".")
	if len(parts) != 3 {
		return false
	}
	var claims map[string]any
	raw, _ := base64.RawURLEncoding.DecodeString(parts[1])
	_ = json.Unmarshal(raw, &claims)
	if claims["iss"] != clientID || claims["sub"] != clientID || claims["aud"] != c.srv.URL+"/auth/realms/netsocs" {
		t.Errorf("assertion claims %v", claims)
		return false
	}
	sig, _ := base64.RawURLEncoding.DecodeString(parts[2])
	if len(sig) != 64 {
		return false
	}
	sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	return ecdsa.Verify(c.key, sum[:], new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:]))
}

func (c *central) hub() string { return c.srv.URL + "/api/netsocs/dh" }

func (c *central) cfg() *Config {
	return &Config{ClientID: "drv-" + strings.Repeat("a", 32), TokenURL: c.srv.URL + "/auth/realms/netsocs/protocol/openid-connect/token", KeyFile: "driver.key"}
}

func (c *central) callHub(t *testing.T) http.Header {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, c.hub()+"/objects", nil)
	req.Header.Set("Authorization", driverKey)
	req.Header.Set("X-Auth-Token", "site-token")
	resp, err := httpx.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	return c.lastHub.Load().(http.Header)
}

func TestSetupJWTMode(t *testing.T) {
	t.Cleanup(func() { Activate(nil) })
	dir := t.TempDir()
	c := newCentral(t, &writeKey(t, dir).PublicKey)
	atomic.StoreInt32(&c.tokenFail, 2) // Keycloak still starting: backoff, no exit

	mode, p, err := Setup(context.Background(), c.cfg(), dir, driverKey, c.hub(), t.Logf)
	if err != nil || mode != ModeJWT || p == nil {
		t.Fatalf("mode %s err %v", mode, err)
	}
	h := c.callHub(t)
	if h.Get("Authorization") != "Bearer jwt-1" || h.Get("X-Driver-Key") != driverKey || h.Get("X-Auth-Token") != "" {
		t.Fatalf("DriverHub got Authorization=%q X-Driver-Key=%q X-Auth-Token=%q",
			h.Get("Authorization"), h.Get("X-Driver-Key"), h.Get("X-Auth-Token"))
	}
	if !readState(dir, c.hub()).JWTSeen {
		t.Fatal("JWT use not remembered")
	}
}

func TestNoAuthBlockChangesNothing(t *testing.T) {
	t.Cleanup(func() { Activate(nil) })
	c := newCentral(t, nil)
	mode, p, err := Setup(context.Background(), nil, t.TempDir(), driverKey, c.hub(), t.Logf)
	if err != nil || mode != ModeLegacy || p != nil {
		t.Fatalf("mode %s p %v err %v", mode, p, err)
	}
	h := c.callHub(t)
	if h.Get("Authorization") != driverKey || h.Get("X-Auth-Token") != "site-token" || h.Get("X-Driver-Key") != "" {
		t.Fatalf("headers changed: %v", h)
	}
}

func TestOldCentralUsesTheDriverKeyUntilJWTWasSeen(t *testing.T) {
	t.Cleanup(func() { Activate(nil) })
	dir := t.TempDir()
	c := newCentral(t, &writeKey(t, dir).PublicKey)
	atomic.StoreInt32(&c.caps, 1)

	mode, _, err := Setup(context.Background(), c.cfg(), dir, driverKey, c.hub(), t.Logf)
	if err != nil || mode != ModeLegacy {
		t.Fatalf("first contact with an old central: mode %s err %v", mode, err)
	}
	if h := c.callHub(t); h.Get("Authorization") != driverKey {
		t.Fatalf("legacy headers changed: %v", h)
	}

	// Once JWTs were used with this central, a 404 is a downgrade.
	if err := writeState(dir, state{JWTSeen: true, Hub: c.hub()}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Setup(context.Background(), c.cfg(), dir, driverKey, c.hub(), t.Logf); !errors.Is(err, ErrDowngrade) {
		t.Fatalf("downgrade not refused: %v", err)
	}
	// The memory is per central: another hub address starts fresh.
	if s := readState(dir, "https://other/api/netsocs/dh"); s.JWTSeen {
		t.Fatal("state leaked to another central")
	}
}

func TestDiscoveryErrorsAreRetriedNotTakenForAnOldCentral(t *testing.T) {
	t.Cleanup(func() { Activate(nil) })
	dir := t.TempDir()
	c := newCentral(t, &writeKey(t, dir).PublicKey)
	atomic.StoreInt32(&c.caps, 2)
	go func() {
		time.Sleep(1500 * time.Millisecond)
		atomic.StoreInt32(&c.caps, 0)
	}()
	mode, _, err := Setup(context.Background(), c.cfg(), dir, driverKey, c.hub(), t.Logf)
	if err != nil || mode != ModeJWT {
		t.Fatalf("mode %s err %v", mode, err)
	}
}

func TestRejectedIdentityIsFinal(t *testing.T) {
	t.Cleanup(func() { Activate(nil) })
	dir := t.TempDir()
	c := newCentral(t, &writeKey(t, dir).PublicKey)
	c.rejected.Store(true)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, _, err := Setup(ctx, c.cfg(), dir, driverKey, c.hub(), t.Logf); !errors.Is(err, ErrRejected) {
		t.Fatalf("want ErrRejected, got %v", err)
	}
	if Active() != nil {
		t.Fatal("identity activated after a rejection")
	}
}

func TestWrongKeyIsRejected(t *testing.T) {
	dir := t.TempDir()
	writeKey(t, dir)
	other, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	c := newCentral(t, &other.PublicKey)
	p, err := NewProvider(*c.cfg(), dir, driverKey, c.hub())
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Fetch(context.Background()); !errors.Is(err, ErrRejected) {
		t.Fatalf("want ErrRejected, got %v", err)
	}
}

func TestApplyOnlyTouchesTheDriverHub(t *testing.T) {
	dir := t.TempDir()
	writeKey(t, dir)
	p, err := NewProvider(Config{ClientID: "drv-x", TokenURL: "https://c/auth/realms/netsocs/protocol/openid-connect/token", KeyFile: "driver.key"},
		dir, driverKey, "https://central.example/api/netsocs/dh")
	if err != nil {
		t.Fatal(err)
	}
	p.token = "jwt"
	for raw, want := range map[string]bool{
		"https://central.example/api/netsocs/dh/objects":         true,
		"wss://central.example:443/api/netsocs/dh/objects/ws":    true,
		"https://central.example/auth/realms/netsocs/protocol/x": false,
		"https://central.example:8443/api/netsocs/dh/objects":    false,
		"http://192.168.0.12/stw-cgi/system.cgi":                 false,
		"https://evil.example/api/netsocs/dh/objects":            false,
	} {
		u, _ := url.Parse(raw)
		h := http.Header{"Authorization": {driverKey}}
		p.Apply(u, h)
		if got := h.Get("Authorization") == "Bearer jwt"; got != want {
			t.Errorf("%s: applied=%v want %v", raw, got, want)
		}
	}
}

func TestRenewAt(t *testing.T) {
	now := time.Now()
	if got := renewAt(now, now.Add(5*time.Minute)); !got.Equal(now.Add(4 * time.Minute)) {
		t.Errorf("5 min token renews at %v", got.Sub(now))
	}
	if got := renewAt(now, now.Add(90*time.Second)); !got.Equal(now.Add(60 * time.Second)) {
		t.Errorf("90 s token renews at %v", got.Sub(now))
	}
}

func TestKeepFreshHandsOutNewTokens(t *testing.T) {
	dir := t.TempDir()
	c := newCentral(t, &writeKey(t, dir).PublicKey)
	p, _ := NewProvider(*c.cfg(), dir, driverKey, c.hub())
	if err := p.Fetch(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.mu.Lock()
	p.exp = time.Now().Add(300 * time.Millisecond)
	p.mu.Unlock()
	got := make(chan string, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go p.KeepFresh(ctx, func(tok string) {
		select {
		case got <- tok:
		default:
		}
	}, t.Logf)
	select {
	case tok := <-got:
		if tok != "jwt-2" {
			t.Fatalf("renewed token %q", tok)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("token not renewed")
	}
}

func TestExitCodes(t *testing.T) {
	for code, want := range map[int]int{4000: ExitRestart, 4001: ExitTokenExpired, 4003: ExitRevoked} {
		if got, ok := ExitCodeForClose(code); !ok || got != want {
			t.Errorf("close %d -> %d %v", code, got, ok)
		}
	}
	if _, ok := ExitCodeForClose(1006); ok {
		t.Error("abnormal closure mapped to a reason")
	}
}
