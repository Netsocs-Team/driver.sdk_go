package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Netsocs-Team/driver.sdk_go/pkg/httpx"
)

// Capabilities answered by DriverHub on GET /auth/capabilities.
type Capabilities struct {
	DriverJWT  bool `json:"driver_jwt"`
	RequireJWT bool `json:"require_jwt"`
	Reauth     bool `json:"reauth"`
}

// errOldCentral: the DriverHub has no /auth/capabilities (404) or says it
// does not accept JWTs.
var errOldCentral = errors.New("central does not accept driver JWTs")

// ErrDowngrade: this driver used JWTs with this central before and now the
// central says it does not accept them. It may be a rollback of the central
// or someone in the middle answering 404 to force the old mode (possible
// when the auth block has no ca_mode, since TLS is then not verified): it
// needs a person to look, so the driver stops.
var ErrDowngrade = errors.New("auth: the central stopped accepting driver JWTs after this driver used them; " +
	"if the central was rolled back on purpose, delete " + StateFile + " next to driver.netsocs.json")

// StateFile remembers that the central accepted JWTs (like HSTS).
const StateFile = ".netsocs-auth-state.json"

type state struct {
	JWTSeen  bool      `json:"driver_jwt_seen"`
	Hub      string    `json:"driver_hub_host"`
	LastSeen time.Time `json:"last_seen"`
}

func readState(dir, hub string) state {
	b, err := os.ReadFile(filepath.Join(dir, StateFile))
	if err != nil {
		return state{}
	}
	var s state
	if json.Unmarshal(b, &s) != nil || s.Hub != hub {
		return state{}
	}
	return s
}

func writeState(dir string, s state) error {
	b, _ := json.Marshal(s)
	tmp := filepath.Join(dir, StateFile+".tmp")
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, StateFile))
}

// discover asks the DriverHub what it accepts. A 404 is an old central; any
// other failure (network, 5xx) is returned as is: it says nothing about the
// central's version, so the caller retries instead of choosing a mode.
func discover(ctx context.Context, hub string) (Capabilities, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(hub, "/")+"/auth/capabilities", nil)
	if err != nil {
		return Capabilities{}, err
	}
	c := &http.Client{Timeout: 15 * time.Second, Transport: httpx.RoundTripper()}
	resp, err := c.Do(req)
	if err != nil {
		return Capabilities{}, err
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return Capabilities{}, errOldCentral
	case resp.StatusCode != http.StatusOK:
		return Capabilities{}, fmt.Errorf("GET /auth/capabilities: HTTP %d", resp.StatusCode)
	}
	var caps Capabilities
	if err := json.NewDecoder(resp.Body).Decode(&caps); err != nil {
		return Capabilities{}, fmt.Errorf("GET /auth/capabilities: %w", err)
	}
	if !caps.DriverJWT {
		return caps, errOldCentral
	}
	return caps, nil
}

// Mode the driver runs in.
type Mode string

const (
	ModeLegacy Mode = "legacy" // DriverKey + X-Auth-Token, exactly as before
	ModeJWT    Mode = "jwt"
)

// Setup decides the mode and, for JWT, obtains the first token (retrying with
// backoff) and activates the identity for every SDK request. dir is where
// driver.netsocs.json lives. Without an auth block it does nothing at all.
func Setup(ctx context.Context, cfg *Config, dir, driverKey, driverHubHost string, logf func(string, ...any)) (Mode, *Provider, error) {
	if cfg == nil {
		return ModeLegacy, nil, nil
	}
	if logf == nil {
		logf = func(string, ...any) {}
	}
	p, err := NewProvider(*cfg, dir, driverKey, driverHubHost)
	if err != nil {
		return "", nil, err
	}
	// Before the first request: discovery already goes to the DriverHub.
	if err := secureCentral(*cfg, dir, driverHubHost, logf); err != nil {
		return "", nil, err
	}

	wait := time.Second
	for {
		_, err := discover(ctx, driverHubHost)
		if err == nil {
			break
		}
		if errors.Is(err, errOldCentral) {
			if readState(dir, driverHubHost).JWTSeen {
				return "", nil, ErrDowngrade
			}
			logf("auth: this DriverHub does not accept driver JWTs yet: using the DriverKey (old mode)")
			return ModeLegacy, nil, nil
		}
		logf("auth: cannot ask the DriverHub what it accepts (%v), retrying in %s", err, wait)
		select {
		case <-ctx.Done():
			return "", nil, ctx.Err()
		case <-time.After(wait):
		}
		if wait *= 2; wait > time.Minute {
			wait = time.Minute
		}
	}

	if err := p.FetchWithBackoff(ctx, logf); err != nil {
		return "", nil, err
	}
	if err := writeState(dir, state{JWTSeen: true, Hub: driverHubHost, LastSeen: time.Now().UTC()}); err != nil {
		logf("auth: cannot write %s: %v", StateFile, err)
	}
	Activate(p)
	logf("auth: connected as %s (Keycloak identity)", cfg.ClientID)
	return ModeJWT, p, nil
}

// ErrTLSConfig: the auth block asks for verified TLS (ca_mode) and it cannot
// be applied. The SDK never falls back to unverified TLS in that case.
var ErrTLSConfig = errors.New("auth: unusable TLS settings in the auth block")

// secureCentral applies ca_mode/ca_file (contract C1) to every SDK
// connection to the DriverHub and to the token endpoint. Without ca_mode it
// does nothing: unverified TLS, as in v0.8.0-rc.1.
func secureCentral(cfg Config, dir, driverHubHost string, logf func(string, ...any)) error {
	if cfg.CAMode == "" {
		if cfg.CAFile != "" {
			logf("auth: ca_file without ca_mode is ignored: the central's TLS is not verified")
		}
		return httpx.SecureCentral("", nil)
	}
	var hosts []string
	for _, raw := range []string{driverHubHost, cfg.TokenURL} {
		u, err := url.Parse(raw)
		if err != nil || u.Hostname() == "" {
			return fmt.Errorf("%w: invalid URL %q", ErrTLSConfig, raw)
		}
		if !strings.EqualFold(u.Scheme, "https") {
			return fmt.Errorf("%w: %s must be https:// with ca_mode %q", ErrTLSConfig, raw, cfg.CAMode)
		}
		hosts = append(hosts, u.Hostname())
	}
	var caPEM []byte
	if cfg.CAMode == httpx.CAModePinned {
		if cfg.CAFile == "" {
			return fmt.Errorf("%w: ca_mode %q needs ca_file", ErrTLSConfig, cfg.CAMode)
		}
		path := cfg.CAFile
		if !filepath.IsAbs(path) {
			path = filepath.Join(dir, path)
		}
		var err error
		if caPEM, err = os.ReadFile(filepath.Clean(path)); err != nil {
			return fmt.Errorf("%w: ca_file: %v", ErrTLSConfig, err)
		}
	}
	if err := httpx.SecureCentral(cfg.CAMode, caPEM, hosts...); err != nil {
		return fmt.Errorf("%w: %v", ErrTLSConfig, err)
	}
	logf("auth: TLS of the central verified (ca_mode %s)", cfg.CAMode)
	return nil
}
