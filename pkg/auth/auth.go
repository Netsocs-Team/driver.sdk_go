// Package auth gives a driver its own Keycloak identity (drv-<driver_id>
// client, private_key_jwt), phase 2 of docs/SEGURIDAD-DISENO-AUTH-DRIVERS.md
// (netsocs-docker-compose).
//
// It is opt-in: only a driver.netsocs.json with an "auth" block (written by
// SmartConnect when it enrolls the driver) uses it, and only against a
// DriverHub that says it accepts JWTs (GET /auth/capabilities). Otherwise the
// SDK sends exactly what it always sent (DriverKey, X-Auth-Token).
package auth

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Netsocs-Team/driver.sdk_go/pkg/httpx"
)

// Config is the "auth" block of driver.netsocs.json.
type Config struct {
	ClientID string `json:"client_id"`     // drv-<driver_id>
	TokenURL string `json:"token_url"`     // Keycloak token endpoint of the central
	KeyFile  string `json:"key_file"`      // PEM private key, relative to driver.netsocs.json
	Alg      string `json:"alg,omitempty"` // ES256 (default) or RS256
	DriverID string `json:"driver_id,omitempty"`
	// How to verify the TLS of the central (DriverHub and token_url), contract
	// C1. Absent: unverified, as in v0.8.0-rc.1, for old centrals.
	CAFile string `json:"ca_file,omitempty"` // PEM root CA for "pinned", relative to driver.netsocs.json
	CAMode string `json:"ca_mode,omitempty"` // "pinned" or "public"
}

// Valid reports whether the block is complete.
func (c *Config) Valid() bool {
	return c != nil && strings.HasPrefix(c.ClientID, "drv-") && c.TokenURL != "" && c.KeyFile != ""
}

// Process exit codes of a driver whose configuration socket DriverHub closed
// with a reason (only in JWT mode). The supervisor (SmartConnect, systemd)
// restarts the driver for all of them except ExitRevoked.
const (
	ExitRestart      = 75 // close 4000, or DriverHub lost or not reachable yet: launch the driver again soon (not a crash)
	ExitTokenExpired = 76 // close 4001: token expired or reauth rejected
	ExitRevoked      = 77 // close 4003: identity revoked, do not restart
	ExitDowngrade    = 78 // the central stopped accepting JWTs after this driver used them, or its TLS settings are unusable
)

// Provider holds the driver key and the current access token.
type Provider struct {
	cfg       Config
	signer    crypto.Signer
	alg       string
	driverKey string
	hub       *url.URL
	httpc     *http.Client
	now       func() time.Time

	mu    sync.RWMutex
	token string
	exp   time.Time
}

// NewProvider loads the private key. baseDir resolves a relative key_file.
func NewProvider(cfg Config, baseDir, driverKey, driverHubHost string) (*Provider, error) {
	if !cfg.Valid() {
		return nil, errors.New("auth: incomplete auth block (client_id drv-*, token_url, key_file)")
	}
	keyPath := cfg.KeyFile
	if !filepath.IsAbs(keyPath) {
		keyPath = filepath.Join(baseDir, keyPath)
	}
	pemBytes, err := os.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return nil, fmt.Errorf("auth: driver key: %w", err)
	}
	signer, alg, err := parseKey(pemBytes)
	if err != nil {
		return nil, err
	}
	if cfg.Alg != "" && cfg.Alg != alg {
		return nil, fmt.Errorf("auth: key is %s but the auth block says %s", alg, cfg.Alg)
	}
	hub, err := url.Parse(driverHubHost)
	if err != nil || hub.Host == "" {
		return nil, fmt.Errorf("auth: invalid driver_hub_host %q", driverHubHost)
	}
	return &Provider{
		cfg: cfg, signer: signer, alg: alg, driverKey: driverKey, hub: hub,
		// Not the hooked SDK client: the token request must not carry a token.
		httpc: &http.Client{Timeout: 20 * time.Second, Transport: httpx.RoundTripper()},
		now:   time.Now,
	}, nil
}

func parseKey(pemBytes []byte) (crypto.Signer, string, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, "", errors.New("auth: driver key is not PEM")
	}
	var key any
	var err error
	switch block.Type {
	case "EC PRIVATE KEY":
		key, err = x509.ParseECPrivateKey(block.Bytes)
	case "RSA PRIVATE KEY":
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	default:
		key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	}
	if err != nil {
		return nil, "", fmt.Errorf("auth: driver key: %w", err)
	}
	switch k := key.(type) {
	case *ecdsa.PrivateKey:
		if k.Curve.Params().Name != "P-256" {
			return nil, "", errors.New("auth: EC driver key must be P-256")
		}
		return k, "ES256", nil
	case *rsa.PrivateKey:
		if k.N.BitLen() < 2048 {
			return nil, "", errors.New("auth: RSA driver key must be >= 2048 bits")
		}
		return k, "RS256", nil
	}
	return nil, "", errors.New("auth: unsupported driver key type")
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

// assertion is the client_assertion of private_key_jwt (RFC 7523): signed by
// the driver key, for the realm (audience = issuer), valid one minute.
func (p *Provider) assertion() (string, error) {
	now := p.now()
	jti := make([]byte, 16)
	_, _ = rand.Read(jti)
	header, _ := json.Marshal(map[string]string{"alg": p.alg, "typ": "JWT"})
	claims, _ := json.Marshal(map[string]any{
		"iss": p.cfg.ClientID, "sub": p.cfg.ClientID,
		"aud": strings.TrimSuffix(p.cfg.TokenURL, "/protocol/openid-connect/token"),
		"jti": b64(jti), "iat": now.Unix(), "exp": now.Add(time.Minute).Unix(),
	})
	signing := b64(header) + "." + b64(claims)
	sum := sha256.Sum256([]byte(signing))
	var sig []byte
	switch k := p.signer.(type) {
	case *ecdsa.PrivateKey:
		r, s, err := ecdsa.Sign(rand.Reader, k, sum[:])
		if err != nil {
			return "", err
		}
		sig = append(r.FillBytes(make([]byte, 32)), s.FillBytes(make([]byte, 32))...)
	case *rsa.PrivateKey:
		var err error
		if sig, err = rsa.SignPKCS1v15(rand.Reader, k, crypto.SHA256, sum[:]); err != nil {
			return "", err
		}
	}
	return signing + "." + b64(sig), nil
}

// ErrRejected: Keycloak refused the client (deleted, disabled or wrong key).
// Retrying does not help.
var ErrRejected = errors.New("auth: Keycloak rejected the driver identity")

// Fetch requests a new access token (client_credentials + private_key_jwt).
func (p *Provider) Fetch(ctx context.Context) error {
	a, err := p.assertion()
	if err != nil {
		return err
	}
	form := url.Values{
		"grant_type":            {"client_credentials"},
		"client_id":             {p.cfg.ClientID},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {a},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := p.httpc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("%w: HTTP %d %s", ErrRejected, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth: token endpoint HTTP %d", resp.StatusCode)
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tok); err != nil || tok.AccessToken == "" {
		return errors.New("auth: token endpoint: invalid response")
	}
	exp := p.now().Add(time.Duration(tok.ExpiresIn) * time.Second)
	p.mu.Lock()
	p.token, p.exp = tok.AccessToken, exp
	p.mu.Unlock()
	return nil
}

// FetchWithBackoff retries until a token is obtained or ctx ends, so a
// Keycloak that is down or unreachable does not turn into a restart loop of
// every driver (section 9 of the design). A rejected identity is final.
func (p *Provider) FetchWithBackoff(ctx context.Context, logf func(string, ...any)) error {
	wait := time.Second
	for {
		err := p.Fetch(ctx)
		if err == nil {
			return nil
		}
		if errors.Is(err, ErrRejected) {
			return err
		}
		if logf != nil {
			logf("auth: no token yet (%v), retrying in %s", err, wait)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
		if wait *= 2; wait > time.Minute {
			wait = time.Minute
		}
	}
}

// Token is the current access token.
func (p *Provider) Token() (string, time.Time) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.token, p.exp
}

// renewAt is when to fetch the next token: a minute before expiry, or at
// two thirds of a short lifetime.
func renewAt(now, exp time.Time) time.Time {
	life := exp.Sub(now)
	lead := time.Minute
	if life < 3*time.Minute {
		lead = life / 3
	}
	return exp.Add(-lead)
}

// KeepFresh renews the token before it expires and hands every new one to
// onNew (the configuration socket sends it as a reauth message). If renewal
// keeps failing the token simply expires and DriverHub closes the socket
// (4001), which ends the process as it must.
func (p *Provider) KeepFresh(ctx context.Context, onNew func(token string), logf func(string, ...any)) {
	for {
		_, exp := p.Token()
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Until(renewAt(p.now(), exp))):
		}
		for {
			err := p.Fetch(ctx)
			if err == nil {
				break
			}
			if logf != nil {
				logf("auth: token renewal failed: %v", err)
			}
			if errors.Is(err, ErrRejected) {
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
			}
		}
		tok, _ := p.Token()
		if onNew != nil {
			onNew(tok)
		}
	}
}

// Apply sets the credentials of a request to the DriverHub: the bearer token
// and the DriverKey (driver type metadata) in X-Driver-Key. Requests to any
// other host or path (cameras, Keycloak) are untouched.
func (p *Provider) Apply(u *url.URL, h http.Header) {
	if u == nil || !strings.EqualFold(hostPort(u), hostPort(p.hub)) ||
		!strings.HasPrefix(u.Path, strings.TrimRight(p.hub.Path, "/")) {
		return
	}
	tok, _ := p.Token()
	if tok == "" {
		return
	}
	h.Set("Authorization", "Bearer "+tok)
	h.Set("X-Driver-Key", p.driverKey)
	h.Del("X-Auth-Token")
}

func hostPort(u *url.URL) string {
	port := u.Port()
	if port == "" {
		switch u.Scheme {
		case "https", "wss":
			port = "443"
		default:
			port = "80"
		}
	}
	return u.Hostname() + ":" + port
}

// ------------------------------------------------------------ process-wide

var active struct {
	sync.RWMutex
	p *Provider
}

// Active is the provider in use, nil in the old mode.
func Active() *Provider {
	active.RLock()
	defer active.RUnlock()
	return active.p
}

// Activate makes p the identity of every SDK request to the DriverHub.
func Activate(p *Provider) {
	active.Lock()
	active.p = p
	active.Unlock()
	if p == nil {
		httpx.SetAuthHook(nil)
		return
	}
	httpx.SetAuthHook(p.Apply)
}

// ExitCodeForClose maps the close code of the configuration socket to the
// exit code of the process; ok is false for codes DriverHub does not set.
func ExitCodeForClose(code int) (exit int, ok bool) {
	switch code {
	case 4000:
		return ExitRestart, true
	case 4001:
		return ExitTokenExpired, true
	case 4003:
		return ExitRevoked, true
	}
	return 0, false
}
